package launcher

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// LaunchRemoteManager launches the manager in kubernetes cluster, that will instantiate jobs and manage them until the end
func LaunchRemoteManager(config *viper.Viper, specsDirectory string, duration time.Duration, shouldReportMetrics, shouldDeleteAllResources bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchRemoteManager",
	})

	specs, err := GetSpecs(specsDirectory)
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
		pbKubernetes.DeleteAllManager(logger, clientset, config)
		pbKubernetes.DeleteAll(logger, clientset, config)
		return
	}

	pbKubernetes.DeployJobsRemote(logger, clientset, config, specs, duration)
	controller := pbKubernetes.NewManagerController(logger, clientset, config)
	controller.Run(1, duration)
	pbKubernetes.DeleteAll(logger, clientset, config)
	pbKubernetes.DeleteAllManager(logger, clientset, config)
}
