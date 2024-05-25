package log

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_SetUpLogrus(t *testing.T) {
	t.Run("should set logrus level to info", func(t *testing.T) {
		err := SetUpLogrus("info")
		assert.Nil(t, err)
		assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	})
	t.Run("should set logrus level to debug", func(t *testing.T) {
		err := SetUpLogrus("debug")
		assert.Nil(t, err)
		assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	})
	t.Run("should return error when invalid logrus level", func(t *testing.T) {
		err := SetUpLogrus("invalid")
		assert.NotNil(t, err)
	})
	t.Run("should setup text formatter", func(t *testing.T) {
		err := SetUpLogrus("info")
		assert.Nil(t, err)
		assert.IsType(t, &logrus.TextFormatter{}, logrus.StandardLogger().Formatter)

		// Check formatter timestamp format
		testTextFormater := &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		}
		assert.Equal(t, testTextFormater, logrus.StandardLogger().Formatter)
	})
}
