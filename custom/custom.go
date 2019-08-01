package custom

import (
	"github.com/topfreegames/pitaya-bot/custom/redis"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/storage"

	"github.com/spf13/viper"
)

// Valid pre function types
const (
	PreRunFunctionRedis = "redis"
)

// PreOperation is the interface that structs must implement for preRun
type PreOperation interface {
	Run(args map[string]interface{}) (storage.Storage, error)
}

// PostOperation is the interface that structs must implement for postRun
type PostOperation interface {
	Run(args map[string]interface{}, store storage.Storage) error
}

// DummyPre does nothing
type DummyPre struct{}

// Run returns an empty storage
func (d *DummyPre) Run(args map[string]interface{}) (storage.Storage, error) {
	return &storage.MemoryStorage{}, nil
}

// GetPre returns the pre operation for the spec, if it exists
func GetPre(config *viper.Viper, spec *models.Spec) PreOperation {
	if spec.PreRun == nil {
		return &DummyPre{}
	}

	switch spec.PreRun.Function {
	case PreRunFunctionRedis:
		return redis.GetPre(config)
	}
	return &DummyPre{}
}
