package storage

import (
	"encoding/json"

	"github.com/topfreegames/pitaya-bot/constants"
)

// MemoryStorage is the in memory storage implementation
type MemoryStorage map[string]interface{}

// NewMemoryStorage returns a new MemoryStorage from map
func NewMemoryStorage(m map[string]interface{}) *MemoryStorage {
	if m == nil {
		m = make(map[string]interface{})
	}
	mem := MemoryStorage(m)
	return &mem
}

// Get returns value from key
func (s *MemoryStorage) Get(key string) (interface{}, error) {
	i := map[string]interface{}(*s)
	v, ok := i[key]
	if !ok {
		return nil, constants.ErrStorageKeyNotFound
	}
	return v, nil
}

// Set saves the key and value
func (s *MemoryStorage) Set(key string, val interface{}) error {
	i := map[string]interface{}(*s)
	i[key] = val
	return nil
}

func (s MemoryStorage) String() string {
	j, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(j)
}
