package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/constants"
)

func TestMemoryStorageGet(t *testing.T) {
	var testMemoryStorageGetTable = map[string]struct {
		store  Storage
		result interface{}
		err    error
	}{
		"success_bool":   {&MemoryStorage{"attr": true}, true, nil},
		"success_float":  {&MemoryStorage{"attr": 123.456}, 123.456, nil},
		"success_string": {&MemoryStorage{"attr": "ok"}, "ok", nil},
		"err_not_found":  {&MemoryStorage{}, nil, constants.ErrStorageKeyNotFound},
	}

	for name, table := range testMemoryStorageGetTable {
		t.Run(name, func(t *testing.T) {
			result, err := table.store.Get("attr")
			assert.Equal(t, table.result, result)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestMemoryStorageSet(t *testing.T) {
	var testMemoryStorageGetTable = map[string]struct {
		value interface{}
	}{
		"success_bool":   {true},
		"success_float":  {123.456},
		"success_string": {"ok"},
	}

	for name, table := range testMemoryStorageGetTable {
		t.Run(name, func(t *testing.T) {
			store := &MemoryStorage{}
			err := store.Set("attr", table.value)
			assert.NoError(t, err)
			result, err := store.Get("attr")
			assert.Equal(t, table.value, result)
			assert.NoError(t, err)
		})
	}
}
