package messaging

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(redisAddr string) (*RedisPublisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &RedisPublisher{client: client}, nil
}

func (p *RedisPublisher) Publish(ctx context.Context, channel string, message []byte) error {
	return p.client.Publish(ctx, channel, message).Err()
}

func (p *RedisPublisher) Close() error {
	return p.client.Close()
}

