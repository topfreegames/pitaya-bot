package kubernetes

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateManagerPod will deploy a kubernetes pod containing a pitaya-bot manager
func CreateManagerPod(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper, specs []*models.Spec, duration time.Duration, shouldReportMetrics bool) {
	deploymentsClient := clientset.AppsV1().Deployments(config.GetString("kubernetes.namespace"))
	app := "pitaya-bot-manager"
	if configMapExist(app, logger, clientset, config) {
		return
	}

	configBinary, err := ioutil.ReadFile(config.ConfigFileUsed())
	if err != nil {
		logger.Fatal(err)
	}
	managerConfig := kubernetesAcceptedNamespace(fmt.Sprintf("%s-manager-config", config.GetString("game")))
	createConfigMap(managerConfig, app, map[string][]byte{"config.yaml": configBinary}, logger, clientset, config)

	binData := make(map[string][]byte, len(specs))
	for _, spec := range specs {
		specBinary, err := ioutil.ReadFile(spec.Name)
		if err != nil {
			logger.Fatal(err)
		}
		binData[filepath.Base(spec.Name)] = specBinary
	}
	managerSpecs := kubernetesAcceptedNamespace(fmt.Sprintf("%s-manager-specs", config.GetString("game")))
	createConfigMap(managerSpecs, app, binData, logger, clientset, config)

	managerName := kubernetesAcceptedNamespace(fmt.Sprintf("%s-manager", config.GetString("game")))

	deployment := &appsv1.Deployment{
		ObjectMeta: newObjectMeta(managerName, app, config),
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  app,
					"game": config.GetString("game"),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: newObjectMeta("pod", app, config),
				Spec:       newJobSpec(corev1.RestartPolicyAlways, managerSpecs, managerConfig, "remote-manager", duration, shouldReportMetrics, config),
			},
		},
	}

	if _, err := deploymentsClient.Create(deployment); err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Created manager pod")
}

// DeployJobsRemote will deploy as many kubernetes jobs as number of spec files from remote
func DeployJobsRemote(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper, specs []*models.Spec, duration time.Duration, shouldReportMetrics bool) {
	if configMapExist("pitaya-bot", logger, clientset, config) {
		return
	}

	deployJobs(logger, clientset, config, specs, duration, shouldReportMetrics)
}

// DeployJobsLocal will deploy as many kubernetes jobs as number of spec files from local
func DeployJobsLocal(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper, specs []*models.Spec, duration time.Duration, shouldReportMetrics bool) {
	if configMapExist("pitaya-bot", logger, clientset, config) || configMapExist("pitaya-bot-manager", logger, clientset, config) {
		return
	}

	deployJobs(logger, clientset, config, specs, duration, shouldReportMetrics)
}

func deployJobs(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper, specs []*models.Spec, duration time.Duration, shouldReportMetrics bool) {
	deploymentsClient := clientset.BatchV1().Jobs(config.GetString("kubernetes.namespace"))
	configBinary, err := ioutil.ReadFile(config.ConfigFileUsed())
	app := "pitaya-bot"
	if err != nil {
		logger.Fatal(err)
	}

	configName := kubernetesAcceptedNamespace(fmt.Sprintf("%s-config", config.GetString("game")))
	createConfigMap(configName, app, map[string][]byte{"config.yaml": configBinary}, logger, clientset, config)

	for _, spec := range specs {
		specBinary, err := ioutil.ReadFile(spec.Name)
		if err != nil {
			logger.Fatal(err)
		}
		specName := kubernetesAcceptedNamespace(fmt.Sprintf("%s-%s", config.GetString("game"), filepath.Base(spec.Name)))
		createConfigMap(specName, app, map[string][]byte{filepath.Base(spec.Name): specBinary}, logger, clientset, config)

		deployment := &batchv1.Job{
			ObjectMeta: newObjectMeta(specName, app, config),
			Spec: batchv1.JobSpec{
				//Parallelism: int32Ptr(1), TODO: Via config file, see how many bots are to be instantiated
				BackoffLimit: int32Ptr(config.GetInt32("kubernetes.job.retry")),
				Completions:  int32Ptr(1),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: newObjectMeta("job", app, config),
					Spec:       newJobSpec(corev1.RestartPolicyNever, specName, configName, "local", duration, shouldReportMetrics, config),
				},
			},
		}

		if _, err := deploymentsClient.Create(deployment); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Created job %s", specName)
	}
}

func newObjectMeta(name, app string, config *viper.Viper) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
		Labels: map[string]string{
			"app":  app,
			"game": config.GetString("game"),
		},
	}
}

func newJobSpec(restartPolicy corev1.RestartPolicy, specName, configName, workflowType string, duration time.Duration, shouldReportMetrics bool, config *viper.Viper) corev1.PodSpec {
	return corev1.PodSpec{
		RestartPolicy: restartPolicy,
		Containers: []corev1.Container{
			{
				ImagePullPolicy: corev1.PullPolicy(config.GetString("kubernetes.imagepull")),
				Name:            specName,
				Image:           config.GetString("kubernetes.image"),
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      specName,
						MountPath: "/etc/pitaya-bot/specs",
					},
					{
						Name:      configName,
						MountPath: "/etc/pitaya-bot",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.Quantity{Format: resource.Format(config.GetString("kubernetes.cpu"))},
						corev1.ResourceMemory: resource.Quantity{Format: resource.Format(config.GetString("kubernetes.memory"))},
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.Quantity{Format: resource.Format(config.GetString("kubernetes.cpu"))},
						corev1.ResourceMemory: resource.Quantity{Format: resource.Format(config.GetString("kubernetes.memory"))},
					},
				},
				Command: []string{"./main"},
				Args:    []string{"run", "--config", "/etc/pitaya-bot/config.yaml", "--duration", duration.String(), "-d", "/etc/pitaya-bot/specs", "-t", workflowType, fmt.Sprintf("--report-metrics=%t", shouldReportMetrics)},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: specName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: specName},
					},
				},
			},
			{
				Name: configName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: configName},
					},
				},
			},
		},
	}
}

func configMapExist(app string, logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper) bool {
	configMaps, err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s,game=%s", app, config.GetString("game"))})
	if err != nil {
		logger.Fatal(err)
	}
	return len(configMaps.Items) > 0
}

func createConfigMap(name, app string, binData map[string][]byte, logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":  app,
				"game": config.GetString("game"),
			},
		},
		BinaryData: binData,
	}

	if _, err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).Create(configMap); err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Created spec configMap => name: %s, app: %s", name, app)
}

// DeleteAll will delete all kubernetes resources that have been allocated to make the jobs
func DeleteAll(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper) {
	deleteAll("pitaya-bot", logger, clientset, config)
}

// DeleteAllManager will delete all pitaya-bot managers that have been allocated inside kubernetes cluster
func DeleteAllManager(logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper) {
	deleteAll("pitaya-bot-manager", logger, clientset, config)
}

func deleteAll(app string, logger logrus.FieldLogger, clientset kubernetes.Interface, config *viper.Viper) {
	err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s,game=%s", app, config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete configMaps")
	}
	logger.Infof("Deleted configMaps")

	err = clientset.BatchV1().Jobs(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s,game=%s", app, config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete jobs")
	}
	logger.Infof("Deleted jobs")

	err = clientset.AppsV1().Deployments(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s,game=%s", app, config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete deployments")
	}
	logger.Infof("Deleted deployments")

	err = clientset.CoreV1().Pods(config.GetString("kubernetes.namespace")).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s,game=%s", app, config.GetString("game"))})
	if err != nil {
		logger.WithError(err).Error("Failed to delete pods")
	}
	logger.Infof("Deleted pods")
}
