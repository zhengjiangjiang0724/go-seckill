package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-seckill/config"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis(cfg *config.Config) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// Get 获取缓存
func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// Set 设置缓存
func Set(key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// Del 删除缓存
func Del(key string) error {
	return RDB.Del(ctx, key).Err()
}

// Incr 递增
func Incr(key string) (int64, error) {
	return RDB.Incr(ctx, key).Result()
}

// Decr 递减
func Decr(key string) (int64, error) {
	return RDB.Decr(ctx, key).Result()
}

// HGet 获取哈希字段
func HGet(key, field string) (string, error) {
	return RDB.HGet(ctx, key, field).Result()
}

// HSet 设置哈希字段
func HSet(key, field string, value interface{}) error {
	return RDB.HSet(ctx, key, field, value).Err()
}

// HIncrBy 哈希字段递增
func HIncrBy(key, field string, incr int64) (int64, error) {
	return RDB.HIncrBy(ctx, key, field, incr).Result()
}

// SetNX 设置键，仅当键不存在时
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return RDB.SetNX(ctx, key, value, expiration).Result()
}

// Eval Lua脚本执行
func Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return RDB.Eval(ctx, script, keys, args...).Result()
}

