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
	log := logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "launch",
	})

	spec, err := readSpec(specPath)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Launching %d bots\n", spec.NumberOfInstances)
	var wg sync.WaitGroup
	var boterr []error
	var errmutex sync.Mutex
	for i := 0; i < spec.NumberOfInstances; i++ {
		wg.Add(1)
		go func(i int) {
			if err := runner.Run(config, spec, i); err != nil {
				errmutex.Lock()
				boterr = append(boterr, err)
				errmutex.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	if len(boterr) > 0 {
		logger.Error(boterr)
		os.Exit(1)
	}
	logger.Info("Finished running bots")
}
