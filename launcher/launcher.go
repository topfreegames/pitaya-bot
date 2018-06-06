package launcher

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/runner"
)

func readSpec(specPath string) (*models.Spec, error) {
	raw, err := ioutil.ReadFile(specPath)
	if err != nil {
		return nil, err
	}

	var spec models.Spec
	err = json.Unmarshal(raw, &spec)
	return &spec, err
}

// Launch launches the bot spec
func Launch(config *viper.Viper, specPath string) {
	spec, err := readSpec(specPath)
	if err != nil {
		panic(err)
	}
	log := logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	log.Out = os.Stdout
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "launch",
	})

	logger.Infof("Launching %d bots\n", spec.NumberOfInstances)
	var wg sync.WaitGroup
	for i := 0; i < spec.NumberOfInstances; i++ {
		wg.Add(1)
		go func(i int) {
			runner.Run(config, spec, i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	logger.Info("Finished running bots")
}
