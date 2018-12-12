package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/logger"
)

func getLogger() logrus.FieldLogger {
	log := logrus.New()

	switch verbose {
	case 0:
		log.Level = logrus.InfoLevel
	case 1:
		log.Level = logrus.WarnLevel
	case 3:
		log.Level = logrus.DebugLevel
	default:
		log.Level = logrus.InfoLevel
	}

	if logJSON {
		log.Formatter = new(logrus.JSONFormatter)
	}

	fieldLogger := log.WithFields(logrus.Fields{
		"app": "pitaya-bot",
	})

	logger.SetLogger(fieldLogger)
	return fieldLogger
}
