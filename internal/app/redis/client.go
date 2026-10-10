package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Client — обёртка над go-redis
type Client struct {
	client *redis.Client
}

// NewClient — создаёт подключение к Redis
func NewClient(ctx context.Context, addr, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // "localhost:6379"
		Password: password, // "password"
		DB:       0,
	})

	// Проверяем связь
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

// Close — закрывает соединение
func (c *Client) Close() error {
	return c.client.Close()
}

// Raw — для доступа к низкоуровневому клиенту (если понадобится)
func (c *Client) Raw() *redis.Client {
	return c.client
}
