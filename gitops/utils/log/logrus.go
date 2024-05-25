package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

func SetUpLogrus(level string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(logrusLevel)

	return nil
}
