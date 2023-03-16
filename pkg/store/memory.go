package store

import (
	"encoding/json"
	"fmt"

	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/util"
)

type MemoryStore struct {
	namespace string
	memory    map[string]string
}

func NewMemoryStore(namespace string) *MemoryStore {
	return &MemoryStore{
		namespace: namespace,
		memory:    make(map[string]string),
	}
}

func (s *MemoryStore) Load(key, value interface{}) error {
	defer util.Timer("memory store load")()
	data, ok := s.memory[cast(key)]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}
	return json.Unmarshal([]byte(data), value)
}

func (s *MemoryStore) Save(key, value interface{}) error {
	defer util.Timer("memory store save")()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.memory[cast(key)] = string(data)
	return nil
}
