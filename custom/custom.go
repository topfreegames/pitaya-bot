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

// Valid post function types
const (
	PostRunFunctionRedis = "redis"
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

// DummyPost does nothing
type DummyPost struct{}

// Run does nothing
func (d *DummyPost) Run(args map[string]interface{}, store storage.Storage) error {
	return nil
}

// GetPre returns the pre operation for the spec, if it exists
func GetPre(config *viper.Viper, spec *models.Spec) (PreOperation, map[string]interface{}) {
	if spec.PreRun == nil {
		return &DummyPre{}, map[string]interface{}{}
	}

	switch spec.PreRun.Function {
	case PreRunFunctionRedis:
		return redis.GetPre(config), spec.PreRun.Args
	}
	return &DummyPre{}, map[string]interface{}{}
}

// GetPost returns the post operation for the spec, if it exists
func GetPost(config *viper.Viper, spec *models.Spec) (PostOperation, map[string]interface{}) {
	if spec.PostRun == nil {
		return &DummyPost{}, map[string]interface{}{}
	}

	switch spec.PostRun.Function {
	case PostRunFunctionRedis:
		return redis.GetPost(config), spec.PostRun.Args
	}
	return &DummyPost{}, map[string]interface{}{}
}
