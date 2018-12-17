package kubernetes

import (
	"fmt"
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// ManagerController represents the pitaya-bot manager kubernetes controller that will be watching all job processes and clean everything up at the end of the whole execution
type ManagerController struct {
	indexer   cache.Indexer
	queue     workqueue.RateLimitingInterface
	informer  cache.Controller
	logger    logrus.FieldLogger
	clientset *kubernetes.Clientset
	config    *viper.Viper
	stopCh    chan struct{}
}

// NewManagerController is the ManagerController constructor
func NewManagerController(logger logrus.FieldLogger, clientset *kubernetes.Clientset, config *viper.Viper) *ManagerController {
	jobListWatcher := cache.NewListWatchFromClient(clientset.BatchV1().RESTClient(), "jobs", config.GetString("kubernetes.namespace"), fields.Everything())
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexer, informer := cache.NewIndexerInformer(jobListWatcher, &batchv1.Job{}, 0, cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	return &ManagerController{
		informer:  informer,
		indexer:   indexer,
		queue:     queue,
		logger:    logger,
		clientset: clientset,
		config:    config,
		stopCh:    make(chan struct{}),
	}
}

// Run is the main loop which the pitaya-bot kubernetes controller will executing
func (c *ManagerController) Run(threadiness int, duration float64) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	c.logger.Infof("Starting pitaya-bot manager controller")

	go c.informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		c.logger.Fatal(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, c.config.GetDuration("manager.wait"), c.stopCh)
	}

	if c.finishedAllJobs() {
		close(c.stopCh)
	}

	fmt.Println(c.getPodElapsedTime())
	go c.printManagerStatus(c.getPodElapsedTime(), duration)

	<-c.stopCh
	c.logger.Infof("Stopping Local Manager Controller")
}

func (c *ManagerController) printManagerStatus(elapsed, duration float64) {
	ticker := time.Tick(500 * time.Millisecond)
	for {
		<-ticker
		elapsed += 0.5
		progress := int(math.Max(math.Min(100.0, (elapsed/duration)*100), 0.0))
		managerStatus := fmt.Sprintf("\rProgress: [%d] [", progress)
		for i := 0; i < 50; i++ {
			if i < progress/2 {
				managerStatus = fmt.Sprintf("%s#", managerStatus)
			} else {
				managerStatus = fmt.Sprintf("%s.", managerStatus)
			}
		}
		managerStatus = fmt.Sprintf("%s]\n\n  JOB                                      | ACTIVE | SUCCESS | FAILED\n+------------------------------------------+--------+---------+--------+\n", managerStatus)
		for _, obj := range c.indexer.List() {
			job := obj.(*batchv1.Job)
			if job.ObjectMeta.Labels["app"] != "pitaya-bot" || job.ObjectMeta.Labels["game"] != c.config.GetString("game") {
				continue
			}
			managerStatus = fmt.Sprintf("%s  %-40s | %-6d | %-7d | %d\n", managerStatus, job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed)
		}
		managerStatus = fmt.Sprintf("%s+------------------------------------------+--------+---------+--------+\n\n", managerStatus)
		fmt.Print(managerStatus)
	}
}

func (c *ManagerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.verifyJobs(key.(string))

	c.handleErr(err, key)
	return true
}

func (c *ManagerController) verifyJobs(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.logger.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		c.logger.Infof("Job %s does not exist anymore", key)
	} else {
		job := obj.(*batchv1.Job)
		c.logger.Infof("Update for Job %s", job.GetName())
	}

	if c.finishedAllJobs() {
		c.logger.Infof("All jobs finished")
		close(c.stopCh)
	}
	return nil
}

func (c *ManagerController) finishedAllJobs() bool {
	for _, obj := range c.indexer.List() {
		job := obj.(*batchv1.Job)
		if job.ObjectMeta.Labels["app"] != "pitaya-bot" || job.ObjectMeta.Labels["game"] != c.config.GetString("game") {
			continue
		}
		if job.Status.Active > 0 || (job.Status.Failed <= *job.Spec.BackoffLimit && job.Status.Succeeded < *job.Spec.Completions) {
			return false
		}
	}
	return true
}

func (c *ManagerController) getPodElapsedTime() float64 {
	for _, obj := range c.indexer.List() {
		job := obj.(*batchv1.Job)
		if job.ObjectMeta.Labels["app"] != "pitaya-bot" || job.ObjectMeta.Labels["game"] != c.config.GetString("game") {
			continue
		}
		return float64(time.Since(job.Status.StartTime.Local())) / float64(time.Second)
	}
	return 0
}

func (c *ManagerController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < c.config.GetInt("manager.maxrequeues") {
		c.logger.Infof("Error syncing job %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	runtime.HandleError(err)
	c.logger.Infof("Dropping job %q out of the queue: %v", key, err)
}

func (c *ManagerController) runWorker() {
	for c.processNextItem() {
	}
}
