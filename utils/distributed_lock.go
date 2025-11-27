package utils

import (
	"fmt"
	"time"

	"go-seckill/cache"
)

// DistributedLock 分布式锁
type DistributedLock struct {
	key        string
	value      string
	expiration time.Duration
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		key:        key,
		value:      fmt.Sprintf("%d", time.Now().UnixNano()),
		expiration: expiration,
	}
}

// Lock 加锁
func (dl *DistributedLock) Lock() (bool, error) {
	return cache.SetNX(dl.key, dl.value, dl.expiration)
}

// Unlock 解锁
func (dl *DistributedLock) Unlock() error {
	// 使用Lua脚本确保原子性：只有持有锁的客户端才能解锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := cache.Eval(script, []string{dl.key}, dl.value)
	return err
}

// TryLockWithRetry 尝试加锁，带重试机制
func (dl *DistributedLock) TryLockWithRetry(maxRetries int, retryDelay time.Duration) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		locked, err := dl.Lock()
		if err != nil {
			return false, err
		}
		if locked {
			return true, nil
		}
		time.Sleep(retryDelay)
	}
	return false, nil
}

