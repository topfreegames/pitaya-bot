package storage

import "github.com/spf13/viper"

// Storage defines the interface which the bots will use to get/set their informations
type Storage interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
}

// NewStorage creates the storage with given config
func NewStorage(config *viper.Viper) (Storage, error) {
	switch config.GetString("storage.type") {
	default:
		return &MemoryStorage{}, nil
	}
}
