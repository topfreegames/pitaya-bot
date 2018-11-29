package bot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/models"
)

func TestError(t *testing.T) {
	t.Run("testExpectError", func(t *testing.T) {
		err := NewExpectError(errors.New("test"), []byte{}, models.ExpectSpec{"$response.code": models.ExpectSpecEntry{Type: "string", Value: "200"}})
		assert.Equal(t, "\nErr: test \nRawData:  \nExpected: {\"$response.code\":{\"type\":\"string\",\"value\":\"200\"}}\n", err.Error())
	})
}
