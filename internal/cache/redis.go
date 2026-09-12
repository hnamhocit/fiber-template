package cache

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/hnamhocit/fiber-template/internal/config"
)

// Client embeds *redis.Client, so all go-redis methods are available as-is.
type Client struct {
	*redis.Client
}

// New builds the Redis client, or (nil, nil) when not configured.
func New(svc config.Services) (*Client, error) {
	if !svc.RedisEnabled() {
		return nil, nil
	}
	opt, err := redis.ParseURL(svc.RedisURL)
	if err != nil {
		return nil, err
	}
	return &Client{Client: redis.NewClient(opt)}, nil
}

// HealthCheck pings the server; used by the readiness probe.
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

// Available reports whether Redis is configured.
// Safe to call on a nil *Client, so features can write `if deps.Redis.Available()`.
func (c *Client) Available() bool { return c != nil }
