package cache

import (
	"context"
	"fmt"
	"myGinServer/config"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	client *redis.Client
}

type StoreCache interface {
	Publish(ctx context.Context, channel string, msg string) error
	Subscribe(ctx context.Context, channel string) (string, error)
}

func NewRedisCache(config *config.CacheConfig) StoreCache {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password:     config.Password,
		PoolSize:     config.PoolSize,
		MinIdleConns: 2,
		DB:           config.DB,
		MaxRetries:   10,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return &RedisStore{client: client}
}

func (r *RedisStore) Publish(ctx context.Context, channel string, msg string) error {
	_, err := r.client.Publish(ctx, channel, msg).Result()
	return err
}

func (r *RedisStore) Subscribe(ctx context.Context, channel string) (string, error) {
	receiveMessage, err := r.client.Subscribe(ctx, channel).ReceiveMessage(ctx)
	if err != nil {
		return "", err
	}
	return receiveMessage.Payload, nil
}
