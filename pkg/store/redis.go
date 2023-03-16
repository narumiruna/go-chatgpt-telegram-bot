package store

import (
	"context"
	"encoding/json"
	"fmt"

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
			Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
			Password: config.Password,
			DB:       config.DB, // use default DB
		}),
		namespace: namespace,
	}
}

func (s *RedisStore) Load(key, value interface{}) error {
	defer util.Timer("redis store load")()
	data, err := s.redis.Get(context.Background(), s.namespace+":"+cast(key)).Result()
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
	return s.redis.Set(context.Background(), s.namespace+":"+cast(key), data, 0).Err()
}
