package launcher

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// LaunchDeleteAll launches the manager in kubernetes cluster, that will instantiate jobs and manage them until the end
func LaunchDeleteAll(config *viper.Viper, logger logrus.FieldLogger) {
	logger = logger.WithFields(logrus.Fields{
		"function": "LaunchDeleteAll",
	})

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.GetString("kubernetes.config")},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: config.GetString("kubernetes.masterurl")}, CurrentContext: config.GetString("kubernetes.context")}).ClientConfig()
	if err != nil {
		logger.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal(err)
	}

	pbKubernetes.DeleteAll(logger, clientset, config)
	pbKubernetes.DeleteAllManager(logger, clientset, config)
}
