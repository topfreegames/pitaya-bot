package storage

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/constants"
)

func TestStorageNewStorage(t *testing.T) {
	t.Parallel()

	tables := map[string]struct {
		typ     string
		resType Storage
		err     error
	}{
		"memory": {
			typ:     "memory",
			resType: &MemoryStorage{},
		},
		"invalid": {
			typ: "invalid",
			err: constants.ErrStorageTypeNotFound,
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			cfg := viper.New()
			cfg.Set("storage.type", table.typ)
			st, err := NewStorage(cfg)
			assert.Equal(t, table.err, err)
			if table.err != nil {
				assert.IsType(t, table.resType, st)
			}
		})
	}
}
