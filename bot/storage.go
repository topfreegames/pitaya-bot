package bot

import (
	"fmt"

	"github.com/spf13/viper"
)

type storage map[string]interface{}

func newStorage(config *viper.Viper) *storage {
	store := storage{}
	return &store
}

func (s *storage) Get(key string) (interface{}, bool) {
	fmt.Println("WILL GET KEY: " + key)
	i := map[string]interface{}(*s)
	v, ok := i[key]
	return v, ok
}

func (s *storage) Set(key string, val interface{}) {
	fmt.Println("WILL SET KEY: " + key)
	i := map[string]interface{}(*s)
	i[key] = val
}
