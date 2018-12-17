package launcher

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// LaunchManagerDeploy launches the deploy to instantiate a pitaya-bot manager pod inside kubernetes
func LaunchManagerDeploy(config *viper.Viper, specsDirectory string, duration time.Duration, shouldReportMetrics, shouldDeleteAllResources bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchManagerDeploy",
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

	pbKubernetes.CreateManagerPod(logger, clientset, config, specs, duration)
}
