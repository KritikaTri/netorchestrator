package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"netorchestrator/internal/config"
)

// Redis represents a Redis cache client
type Redis struct {
	Client *redis.Client
	config config.RedisConfig
	logger *zap.Logger
}

// NewRedis creates a new Redis client
func NewRedis(cfg config.RedisConfig) (*Redis, error) {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisClient := &Redis{
		Client: rdb,
		config: cfg,
		logger: zap.NewNop(), // Will be set later if needed
	}

	return redisClient, nil
}

// SetLogger sets the logger for the Redis client
func (r *Redis) SetLogger(logger *zap.Logger) {
	r.logger = logger
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.Client.Close()
}

// Set sets a key-value pair with expiration
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del deletes a key
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// HSet sets a field in a hash
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash
func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from a hash
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// Health checks the Redis connection health
func (r *Redis) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
