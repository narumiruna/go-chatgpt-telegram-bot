package store

import (
	"context"
	"encoding/json"
	"net"

	"github.com/codingconcepts/env"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/util"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" default:"localhost"`
	Port     string `env:"REDIS_PORT" default:"6379"`
	Password string `env:"REDIS_PASSWORD" default:""`
	DB       int    `env:"REDIS_DB" default:"0"`
}

type RedisStore struct {
	redis     *redis.Client
	namespace string
}

func NewRedisClient(namespace string) *RedisStore {
	var config RedisConfig
	if err := env.Set(&config); err != nil {
		panic(err)
	}

	return &RedisStore{
		redis: redis.NewClient(&redis.Options{
			Addr:     net.JoinHostPort(config.Host, config.Port),
			Password: config.Password,
			DB:       config.DB, // use default DB
		}),
		namespace: namespace,
	}
}
func (s *RedisStore) insertNamespace(key interface{}) string {
	return s.namespace + ":" + cast(key)
}

func (s *RedisStore) Load(key, value interface{}) error {
	defer util.Timer("redis store load")()
	data, err := s.redis.Get(context.Background(), s.insertNamespace(key)).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), value)
}

func (s *RedisStore) Save(key, value interface{}) error {
	defer util.Timer("redis store save")()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.redis.Set(context.Background(), s.insertNamespace(key), data, 0).Err()
}

func (s *RedisStore) Delete(key interface{}) error {
	defer util.Timer("redis store save")()
	return s.redis.Del(context.Background(), s.insertNamespace(key)).Err()
}
