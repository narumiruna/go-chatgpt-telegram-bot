package store

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type Store interface {
	Load(key, value interface{}) error
	Save(key, value interface{}) error
	Delete(key interface{}) error
}

func New(namespace string) Store {
	storeType := os.Getenv("STORE_TYPE")
	switch storeType {
	case "redis":
		log.Infof("using redis store for %s", namespace)
		return NewRedisClient(namespace)
	default:
		log.Infof("using memory store for %s", namespace)
		return NewMemoryStore(namespace)
	}
}
