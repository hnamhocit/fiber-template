package server

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const storagePrefix = "ftpl:limiter:"

type redisStorage struct {
	c *redis.Client
}

func newRedisStorage(c *redis.Client) *redisStorage {
	return &redisStorage{c: c}
}

// Non-context methods (legacy, kept for compatibility)
func (s *redisStorage) Get(key string) ([]byte, error) {
	return s.GetWithContext(context.Background(), key)
}

func (s *redisStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.SetWithContext(context.Background(), key, val, exp)
}

func (s *redisStorage) Delete(key string) error {
	return s.DeleteWithContext(context.Background(), key)
}

func (s *redisStorage) Reset() error {
	return s.ResetWithContext(context.Background())
}

// Context-aware methods (Fiber v3 Storage interface)
func (s *redisStorage) GetWithContext(ctx context.Context, key string) ([]byte, error) {
	val, err := s.c.Get(ctx, storagePrefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return val, err
}

func (s *redisStorage) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error {
	return s.c.Set(ctx, storagePrefix+key, val, exp).Err()
}

func (s *redisStorage) DeleteWithContext(ctx context.Context, key string) error {
	return s.c.Del(ctx, storagePrefix+key).Err()
}

func (s *redisStorage) ResetWithContext(ctx context.Context) error {
	var keys []string
	iter := s.c.Scan(ctx, 0, storagePrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.c.Del(ctx, keys...).Err()
	}
	return nil
}

func (s *redisStorage) Close() error {
	return nil // client owned by bootstrap.Infra
}
