package launcher

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
)

// LaunchLocalManager launches the manager locally, that will instantiate jobs and manage them until the end
func LaunchLocalManager(config *viper.Viper, specsDirectory string, duration time.Duration, shouldReportMetrics bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchLocalManager",
	})

	specs, err := GetSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	clientset := newKubernetesClientset(config, logger)

	pbKubernetes.DeployJobsLocal(logger, clientset, config, specs, duration, shouldReportMetrics)
	controller := pbKubernetes.NewManagerController(logger, clientset, config)
	controller.Run(1, duration)
	pbKubernetes.DeleteAll(logger, clientset, config)
}
