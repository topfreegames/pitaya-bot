package redis

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/topfreegames/pitaya-bot/storage"

	goredis "github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis"
)

// Post defines the post struct for a redis implementation, it receives a storage
type Post struct {
	client *redis.Client
	script *goredis.Script
}

var defaultPostScript = `
redis.replicate_commands()
-- This default script sets a value to a set KEYS[1]..available
local function build_storage(func_key, val)
	redis.call('SADD', func_key..':available', val)
	return "1"
end
-- KEYS[1] is the name received by the Post method
-- ARGV[1] is the value received to be stored
return build_storage(KEYS[1], ARGV[1])
`

var (
	post     *Post
	oncePost sync.Once
)

// GetPost returns a Post instance
func GetPost(
	config *viper.Viper,
) *Post {
	oncePost.Do(func() {
		post = NewPost(config)
	})
	return post
}

// NewPost returns a new Post instance
func NewPost(
	config *viper.Viper,
) *Post {
	client, err := getRedis("custom.redis.post", config)
	if err != nil {
		panic(fmt.Sprintf("failed to get post redis client: %v", err))
	}
	scriptPath := config.GetString("custom.redis.post.script")
	script := defaultPostScript
	if scriptPath != "" {
		scriptBytes, err := ioutil.ReadFile(scriptPath)
		if err != nil {
			panic(fmt.Sprintf("failed to read post script: %v", err))
		}
		script = string(scriptBytes)
	}
	return &Post{
		client: client,
		script: goredis.NewScript(script),
	}
}

// Run runs the configured script and returns the storage
func (p *Post) Run(args map[string]interface{}, store storage.Storage) error {
	var name string
	if nameInt, ok := args["name"]; ok {
		name, ok = nameInt.(string)
		if !ok {
			return fmt.Errorf("invalid type for name")
		}
	}
	if name == "" {
		return fmt.Errorf("missing name")
	}

	err := p.script.Run(p.client.Client, []string{name}, store.String()).Err()
	if err != nil {
		return err
	}
	return nil
}
