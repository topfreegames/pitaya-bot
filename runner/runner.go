package runner

import (
	"errors"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pbot "github.com/topfreegames/pitaya-bot/bot"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/state"
)

// Run runs a bot according to the spec
func Run(app *state.App, config *viper.Viper, spec *models.Spec, id int, log logrus.FieldLogger) error {
	logger := log.WithFields(logrus.Fields{
		"source":   "pitaya-bot",
		"function": "run",
		"botId":    id,
	})

	var err error
	defer func() {
		err := recover()
		if err != nil {
			logger.Error("PANIC")
			logger.Errorf("%s", debug.Stack())

			logger.Error(err)
		}
	}()

	var bot pbot.Bot
	logger.Infof("Starting bot with id: %d", id)
	if spec.SequentialOperations != nil {
		logger.Debug("Found sequential operations")
		bot, err = pbot.NewSequentialBot(config, spec, id, app.MetricsReporter, logger)
		if err != nil {
			logger.WithError(err).Error("Failed to create bot")
			return err
		}
	}

	if bot == nil {
		err := errors.New("No bot types defined")
		logger.Error(err)
		return err
	}

	err = bot.Initialize()
	if err != nil {
		logger.WithError(err).Error("Failed to initialize bot")
		return err
	}

	err = bot.Run()
	if err != nil {
		return err
	}

	err = bot.Finalize()
	if err != nil {
		logger.WithError(err).Error("Failed to finalize bot")
		return err
	}

	logger.Debug("Finished running")

	return err
}
