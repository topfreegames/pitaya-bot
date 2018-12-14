package launcher

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LaunchRemoteManager launches the manager in kubernetes cluster, that will instantiate jobs and manage them until the end
func LaunchRemoteManager(config *viper.Viper, specsDirectory string, duration float64, shouldReportMetrics, shouldDeleteAllResources bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchRemoteManager",
	})

	specs, err := getSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Kubernetes In Cluster Client created")

	if shouldDeleteAllResources {
		pbKubernetes.DeleteAll(logger, clientset, config)
		return
	}

	pbKubernetes.DeployJobs(logger, clientset, config, specs)
	controller := pbKubernetes.NewManagerController(logger, clientset, config)
	controller.Run(1, duration)
	pbKubernetes.DeleteAll(logger, clientset, config)
	pbKubernetes.DeleteAllManager(logger, clientset, config)
}
