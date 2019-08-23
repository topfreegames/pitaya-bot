package redis

import (
	"sync"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis"
)

var (
	clients    map[string]*redis.Client
	redisMutex sync.Mutex
)

func getRedis(
	prefix string,
	config *viper.Viper,
) (*redis.Client, error) {
	redisMutex.Lock()
	defer redisMutex.Unlock()
	if clients == nil {
		clients = make(map[string]*redis.Client)
	}
	if _, ok := clients[prefix]; !ok {
		redis, err := redis.NewClient(prefix, config)
		if err != nil {
			return nil, err
		}
		clients[prefix] = redis
	}

	return clients[prefix], nil
}
