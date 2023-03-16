package store

import (
	"encoding/json"
	"fmt"
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
	data, ok := s.memory[cast(key)]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}
	return json.Unmarshal([]byte(data), value)
}

func (s *MemoryStore) Save(key, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.memory[cast(key)] = string(data)
	return nil
}
