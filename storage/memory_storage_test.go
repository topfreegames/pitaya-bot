package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/constants"
)

func TestMemoryStorageGet(t *testing.T) {
	t.Parallel()

	testMemoryStorageGetTable := map[string]struct {
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
	t.Parallel()

	testMemoryStorageGetTable := map[string]struct {
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

func TestMemoryStorageNew(t *testing.T) {
	t.Parallel()

	tables := map[string]struct {
		m     map[string]interface{}
		store Storage
	}{
		"nil": {
			m:     nil,
			store: &MemoryStorage{},
		},
		"val": {
			m:     map[string]interface{}{"attr": "wat", "attr2": false},
			store: &MemoryStorage{"attr": "wat", "attr2": false},
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			s := NewMemoryStorage(table.m)
			assert.Equal(t, table.store, s)
		})
	}
}

func TestMemoryStorageString(t *testing.T) {
	t.Parallel()

	tables := map[string]struct {
		store  Storage
		result string
	}{
		"success": {
			store:  &MemoryStorage{"attr": true},
			result: `{"attr":true}`,
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, table.result, table.store.String())
		})
	}
}
