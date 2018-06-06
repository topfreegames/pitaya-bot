package runner

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbot "github.com/topfreegames/pitaya-bot/bot"
	"github.com/topfreegames/pitaya-bot/models"
)

// Run runs a bot according to the spec
func Run(config *viper.Viper, spec *models.Spec, id int) {
	log := logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	log.Out = os.Stdout
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "run",
		"botId":    id,
	})

	var bot pbot.Bot
	logger.Info("Starting")
	if spec.SequentialOperations != nil {
		logger.Info("Found sequential operations")
		var err error
		bot, err = pbot.NewSequentialBot(config, spec, id)
		if err != nil {
			logger.WithError(err).Error("Failed to create bot")
			return
		}
	}

	if bot == nil {
		logger.Error("No bot types defined")
		return
	}

	err := bot.Initialize()
	if err != nil {
		logger.WithError(err).Error("Failed to initialize bot")
		return
	}

	err = bot.Run()
	if err != nil {
		logger.WithError(err).Error("Error running bot")
		return
	}

	err = bot.Finalize()
	if err != nil {
		logger.WithError(err).Error("Failed to finalize bot")
		return
	}

	logger.Info("Finished running")
}
