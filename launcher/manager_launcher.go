package launcher

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
)

type managerController struct {
	indexer   cache.Indexer
	queue     workqueue.RateLimitingInterface
	informer  cache.Controller
	logger    logrus.FieldLogger
	clientset *kubernetes.Clientset
	config    *viper.Viper
	stopCh    chan struct{}
}

func newManagerController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, logger logrus.FieldLogger, clientset *kubernetes.Clientset, config *viper.Viper, stopCh chan struct{}) *managerController {
	return &managerController{
		informer:  informer,
		indexer:   indexer,
		queue:     queue,
		logger:    logger,
		clientset: clientset,
		config:    config,
		stopCh:    stopCh,
	}
}

func (c *managerController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *managerController) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.logger.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		c.logger.Infof("Job %s does not exist anymore", key)
	} else {
		job := obj.(*batchv1.Job)
		if c.finishedAllJobs() {
			c.logger.Infof("All jobs finished")
			close(c.stopCh)
			return nil
		}

		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		c.logger.Infof("Update for Job %s", job.GetName())
	}
	return nil
}

func (c *managerController) finishedAllJobs() bool {
	for _, obj := range c.indexer.List() {
		job := obj.(*batchv1.Job)
		if job.Status.Active > 0 || (job.Status.Failed <= *job.Spec.BackoffLimit && job.Status.Succeeded < *job.Spec.Completions) {
			return false
		}
	}
	return true
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *managerController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		c.logger.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	c.logger.Infof("Dropping pod %q out of the queue: %v", key, err)
}

func (c *managerController) run(threadiness int) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	c.logger.Infof("Starting pitaya-bot manager controller")

	go c.informer.Run(c.stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, c.stopCh)
	}

	<-c.stopCh
	c.logger.Infof("Stopping Pod controller")

	err := c.clientset.CoreV1().ConfigMaps(c.config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "app=pitaya-bot,game="})
	if err != nil {
		c.logger.WithError(err).Error("Failed to delete configMaps")
	}
	c.logger.Infof("Deleted configMaps")
	err = c.clientset.BatchV1().Jobs(c.config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "app=pitaya-bot,game="})
	if err != nil {
		c.logger.WithError(err).Error("Failed to delete jobs")
	}
	c.logger.Infof("Deleted jobs")
	err = c.clientset.CoreV1().Pods(c.config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "app=pitaya-bot,game="})
	if err != nil {
		c.logger.WithError(err).Error("Failed to delete pods")
	}
	c.logger.Infof("Deleted pods")
}

func (c *managerController) runWorker() {
	for c.processNextItem() {
	}
}

// LaunchManager launches the manager to create the pods to run specs
func LaunchManager(app *state.App, config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchManager",
	})

	specs, err := getSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal(err)
	}

	configMapClient := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace"))
	deploymentsClient := clientset.BatchV1().Jobs(config.GetString("kubernetes.namespace"))

	configBinary, err := ioutil.ReadFile(config.ConfigFileUsed())
	if err != nil {
		logger.Fatal(err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "config",
			Labels: map[string]string{
				"app":  "pitaya-bot",
				"game": config.GetString("game"),
			},
		},
		BinaryData: map[string][]byte{"config.yaml": configBinary},
	}

	if _, err = configMapClient.Create(configMap); err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Created configMap config.yaml")

	for _, spec := range specs {
		specBinary, err := ioutil.ReadFile(spec.Name)
		if err != nil {
			logger.Fatal(err)
		}
		specName := kubernetesAcceptedNamespace(filepath.Base(spec.Name))

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: specName,
				Labels: map[string]string{
					"app":  "pitaya-bot",
					"game": config.GetString("game"),
				},
			},
			BinaryData: map[string][]byte{"spec.json": specBinary},
		}

		if _, err = configMapClient.Create(configMap); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Created spec configMap %s", specName)

		deployment := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name: specName,
			},
			Spec: batchv1.JobSpec{
				//Parallelism: int32Ptr(1), TODO: Via config file, see how many bots are to be instantiated
				BackoffLimit: int32Ptr(config.GetInt32("kubernetes.job.retry")),
				Completions:  int32Ptr(1),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":  "pitaya-bot",
							"game": config.GetString("game"),
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						Containers: []corev1.Container{
							{
								Name:  "pitaya-bot",
								Image: "tfgco/pitaya-bot:latest",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "spec",
										MountPath: "/etc/pitaya-bot/specs",
									},
									{
										Name:      "config",
										MountPath: "/etc/pitaya-bot",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "spec",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: specName},
									},
								},
							},
							{
								Name: "config",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "config"},
									},
								},
							},
						},
					},
				},
			},
		}

		if _, err := deploymentsClient.Create(deployment); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Created pod %s", specName)
	}

	jobListWatcher := cache.NewListWatchFromClient(clientset.BatchV1().RESTClient(), "jobs", config.GetString("kubernetes.namespace"), fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
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

	stop := make(chan struct{})
	controller := newManagerController(queue, indexer, informer, logger, clientset, config, stop)

	controller.run(1)

	os.Exit(0)
}

func int32Ptr(i int32) *int32 { return &i }

func kubernetesAcceptedNamespace(s string) string {
	rs := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '.' || r == '-' {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	return string(rs)
}
