package redis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/topfreegames/pitaya-bot/storage"

	goredis "github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis"
)

// Pre defines the pre struct for a redis implementation, it returns a memory
// storage
type Pre struct {
	client *redis.Client
	script *goredis.Script
}

var defaultPreScript = `
redis.replicate_commands()
-- This default script gets a value at random from a set KEYS[1]..available
-- and moves it to KEYS[1]..used. If the set is empty it returns nil
local function build_storage(func_key)
	local ret = redis.call('SPOP', func_key..':available')
	if ret then
		redis.call('SADD', func_key..':used', ret)
	end
	return ret
end
-- KEYS[1] is the name received by the Pre method
return build_storage(KEYS[1])
`

var (
	pre     *Pre
	oncePre sync.Once
)

// GetPre returns a Pre instance
func GetPre(
	config *viper.Viper,
) *Pre {
	oncePre.Do(func() {
		pre = NewPre(config)
	})
	return pre
}

// NewPre returns a new Pre instance
func NewPre(
	config *viper.Viper,
) *Pre {
	client, err := getRedis("custom.redis.pre", config)
	if err != nil {
		panic(fmt.Sprintf("failed to get pre redis client: %v", err))
	}
	scriptPath := config.GetString("custom.redis.pre.script")
	script := defaultPreScript
	if scriptPath != "" {
		scriptBytes, err := ioutil.ReadFile(scriptPath)
		if err != nil {
			panic(fmt.Sprintf("failed to read pre script: %v", err))
		}
		script = string(scriptBytes)
	}
	return &Pre{
		client: client,
		script: goredis.NewScript(script),
	}
}

// Run runs the configured script and returns the storage
func (p *Pre) Run(args map[string]interface{}) (storage.Storage, error) {
	var name string
	if nameInt, ok := args["name"]; ok {
		name, ok = nameInt.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for name")
		}
	}
	if name == "" {
		return nil, fmt.Errorf("missing name")
	}

	res, err := p.script.Run(p.client.Client, []string{name}).Result()
	if err != nil {
		return nil, err
	}
	if resStr, ok := res.(string); ok {
		m := make(map[string]interface{})
		err = json.Unmarshal([]byte(resStr), &m)
		if err != nil {
			return nil, err
		}
		return storage.NewMemoryStorage(m), nil
	}
	return nil, fmt.Errorf("invalid script response type")
}
