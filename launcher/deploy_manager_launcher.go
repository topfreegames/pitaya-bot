package launcher

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
)

// LaunchManagerDeploy launches the deploy to instantiate a pitaya-bot manager pod inside kubernetes
func LaunchManagerDeploy(config *viper.Viper, specsDirectory string, duration time.Duration, shouldReportMetrics bool, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchManagerDeploy",
	})

	specs, err := GetSpecs(specsDirectory)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d specs to be executed", len(specs))

	clientset := newKubernetesClientset(config, logger)
	pbKubernetes.CreateManagerPod(logger, clientset, config, specs, duration)
}
