package kubernetes_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/cmd"
	pbKubernetes "github.com/topfreegames/pitaya-bot/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewManagerController(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	config := cmd.CreateConfig("../testing/config/config.yaml")
	managerController := pbKubernetes.NewManagerController(logrus.New(), clientset, config)
	assert.NotNil(t, managerController)
}
