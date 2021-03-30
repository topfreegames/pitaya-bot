package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/constants"
)

func TestOperationValidate(t *testing.T) {
	var tables = map[string]struct {
		op  *Operation
		err    error
	}{
		"success_default": {&Operation{Type: "listen", URI: "metagame.someHandler.someRoute"}, nil},

		"err_nil":     {nil, constants.ErrSpecInvalidNil},
		"err_no_type": {&Operation{Type: ""}, constants.ErrSpecInvalidType},
		"err_no_uri":  {&Operation{Type: "listen"}, constants.ErrSpecInvalidURI},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			err := table.op.Validate()
			assert.Equal(t, table.err, err)
		})
	}
}

