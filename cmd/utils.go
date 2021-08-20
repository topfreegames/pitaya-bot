package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2/logger"
	pitayalogrus "github.com/topfreegames/pitaya/v2/logger/logrus"
)

func getLogger() logrus.FieldLogger {
	log := logrus.New()

	switch verbose {
	case 0:
		log.Level = logrus.ErrorLevel
	case 1:
		log.Level = logrus.WarnLevel
	case 2:
		log.Level = logrus.InfoLevel
	case 3:
		log.Level = logrus.DebugLevel
	default:
		log.Level = logrus.DebugLevel
	}

	if logJSON {
		log.Formatter = new(logrus.JSONFormatter)
	}

	fieldLogger := log.WithFields(logrus.Fields{
		"app": "pitaya-bot",
	})

	logger.SetLogger(pitayalogrus.NewWithEntry(fieldLogger))
	return fieldLogger
}
