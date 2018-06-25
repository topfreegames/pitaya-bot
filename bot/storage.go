package bot

import (
	"github.com/spf13/viper"
)

type storage map[string]interface{}

func newStorage(config *viper.Viper) *storage {
	store := storage{}
	return &store
}

func (s *storage) Get(key string) (interface{}, bool) {
	i := map[string]interface{}(*s)
	v, ok := i[key]
	return v, ok
}

func (s *storage) Set(key string, val interface{}) {
	i := map[string]interface{}(*s)
	i[key] = val
}
