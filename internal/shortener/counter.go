package shortener

import (
	"context"
	"fmt"
	"math/big"

	"UrlShortner/internal/util/base62"

	"github.com/redis/go-redis/v9"
)

// CounterClient defines the interface required for counter increment operations.
type CounterClient interface {
	Incr(ctx context.Context, key string) (int64, error)
}

// RedisCounterAdapter wraps *redis.Client to implement CounterClient.
type RedisCounterAdapter struct {
	Client *redis.Client
}

// NewRedisCounterAdapter creates an adapter for *redis.Client.
func NewRedisCounterAdapter(rdb *redis.Client) CounterClient {
	if rdb == nil {
		return nil
	}
	return &RedisCounterAdapter{Client: rdb}
}

func (r *RedisCounterAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// CounterShortenerConfig holds configuration for the counter-based shortener strategy.
type CounterShortenerConfig struct {
	CounterKey   string
	Multiplier   uint64
	XORMask      uint64
	MaxCharLimit int
}

type counterShortener struct {
	client       CounterClient
	counterKey   string
	multiplier   uint64
	xorMask      uint64
	maxCharLimit int
}

// NewCounterShortener constructs a new CounterShortener.
func NewCounterShortener(client CounterClient, cfg CounterShortenerConfig) Shortener {
	return &counterShortener{
		client:       client,
		counterKey:   cfg.CounterKey,
		multiplier:   cfg.Multiplier,
		xorMask:      cfg.XORMask,
		maxCharLimit: cfg.MaxCharLimit,
	}
}

func (c *counterShortener) Shorten(_ string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("redis counter client is not initialized")
	}

	// 1. Atomically increment counter in Redis
	rawID, err := c.client.Incr(context.Background(), c.counterKey)
	if err != nil {
		return "", fmt.Errorf("counter increment failed: %w", err)
	}

	// 2. Obfuscate the integer: (rawID * Multiplier) ^ XORMask
	// This ensures sequential IDs (1, 2, 3...) produce non-sequential Base62 short codes.
	obfuscatedID := (uint64(rawID) * c.multiplier) ^ c.xorMask

	// 3. Base62 Encode the obfuscated integer
	bigID := new(big.Int).SetUint64(obfuscatedID)
	shortCode := base62.EncodeBigInt(bigID, c.maxCharLimit)

	return shortCode, nil
}
