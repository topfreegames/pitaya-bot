package launcher

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
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

func newManagerController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, logger logrus.FieldLogger, clientset *kubernetes.Clientset, config *viper.Viper) *managerController {
	return &managerController{
		informer:  informer,
		indexer:   indexer,
		queue:     queue,
		logger:    logger,
		clientset: clientset,
		config:    config,
		stopCh:    make(chan struct{}),
	}
}

func (c *managerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.verifyJobs(key.(string))

	c.handleErr(err, key)
	return true
}

func (c *managerController) verifyJobs(key string) error {
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

func (c *managerController) finishedAllJobs() bool {
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

func (c *managerController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		c.logger.Infof("Error syncing job %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	runtime.HandleError(err)
	c.logger.Infof("Dropping job %q out of the queue: %v", key, err)
}

func (c *managerController) run(threadiness int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	c.logger.Infof("Starting pitaya-bot manager controller")

	go c.informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		c.logger.Fatal(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, c.stopCh)
	}

	if c.finishedAllJobs() {
		close(c.stopCh)
	}

	<-c.stopCh
	c.logger.Infof("Stopping Local Manager Controller")
	deleteAllKubernetesResources(c.logger, c.config, c.clientset)
}

func deleteAllKubernetesResources(logger logrus.FieldLogger, config *viper.Viper, clientset *kubernetes.Clientset) {
	err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=pitaya-bot,game=%s", config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete configMaps")
	}
	logger.Infof("Deleted configMaps")

	err = clientset.BatchV1().Jobs(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=pitaya-bot,game=%s", config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete jobs")
	}
	logger.Infof("Deleted jobs")

	err = clientset.CoreV1().Pods(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=pitaya-bot,game=%s", config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete pods")
	}
	logger.Infof("Deleted pods")
}

func (c *managerController) runWorker() {
	for c.processNextItem() {
	}
}

// LaunchLocalManager launches the manager locally, that will instantiate jobs and manage them until the end
func LaunchLocalManager(app *state.App, config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics, shouldDeleteAllResources bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchLocalManager",
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

	if shouldDeleteAllResources {
		deleteAllKubernetesResources(logger, config, clientset)
		return
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=pitaya-bot,game=%s", config.GetString("game"))})
	if err != nil {
		logger.Fatal(err)
	}
	if len(configMaps.Items) <= 0 {
		instantiateKubernetesJobs(logger, clientset, config, specs)
	}

	controller := createManagerController(logger, clientset, config)
	controller.run(1)

	return
}

func instantiateKubernetesJobs(logger logrus.FieldLogger, clientset *kubernetes.Clientset, config *viper.Viper, specs []*models.Spec) {
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
		logger.Infof("Created job %s", specName)
	}
}

func createManagerController(logger logrus.FieldLogger, clientset *kubernetes.Clientset, config *viper.Viper) *managerController {
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

	return newManagerController(queue, indexer, informer, logger, clientset, config)
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
