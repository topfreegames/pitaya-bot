package launcher

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// LaunchLocalManager launches the manager locally, that will instantiate jobs and manage them until the end
func LaunchLocalManager(config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics, shouldDeleteAllResources bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchLocalManager",
	})

	specs, err := getSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	kubeConfig, err := clientcmd.BuildConfigFromFlags(config.GetString("kubernetes.masterurl"), config.GetString("kubernetes.config"))
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal(err)
	}

	if shouldDeleteAllResources {
		pbKubernetes.DeleteAll(logger, clientset, config)
		return
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(config.GetString("kubernetes.namespace")).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=pitaya-bot,game=%s", config.GetString("game"))})
	if err != nil {
		logger.Fatal(err)
	}
	if len(configMaps.Items) <= 0 {
		pbKubernetes.DeployJobs(logger, clientset, config, specs)
	}

	controller := pbKubernetes.NewManagerController(logger, clientset, config)
	controller.Run(1)
	pbKubernetes.DeleteAll(logger, clientset, config)
}
