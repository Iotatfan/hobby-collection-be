package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) *RedisCache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GetConfig().Redis.Host, config.GetConfig().Redis.Port),
		Password: config.GetConfig().Redis.Password,
		DB:       config.GetConfig().Redis.Database,
	})
	return &RedisCache{client: redisClient}

}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Println("Set Cache:", key)
	return r.client.Set(ctx, key, b, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	fmt.Println("Get Cache:", key, val)
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	fmt.Println("Delete Cache:", keys)
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	fmt.Println("Delete Cache By Pattern:", pattern)
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
