package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	capacity  int
	tokens    int
	rate      int // 每秒添加的令牌数
	lastCheck time.Time
	mu        sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:  capacity,
		tokens:    capacity,
		rate:      rate,
		lastCheck: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastCheck)
	tokensToAdd := int(elapsed.Seconds() * float64(tb.rate))

	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastCheck = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimiter 限流中间件
var globalLimiter = NewTokenBucket(10000, 1000) // 容量10000，每秒1000个令牌

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLimiter.Allow() {
			c.JSON(429, gin.H{
				"code": 429,
				"msg":  "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserRateLimiter 用户级限流器
type UserRateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	maxReq  int
	window  time.Duration
}

var userLimiter = &UserRateLimiter{
	buckets: make(map[string]*TokenBucket),
	maxReq:  5,
	window:  time.Second,
}

func GetUserBucket(userID string) *TokenBucket {
	userLimiter.mu.Lock()
	defer userLimiter.mu.Unlock()

	bucket, exists := userLimiter.buckets[userID]
	if !exists {
		bucket = NewTokenBucket(userLimiter.maxReq, userLimiter.maxReq)
		userLimiter.buckets[userID] = bucket
	}
	return bucket
}

func UserRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP() // 如果没有user_id，使用IP
		}

		bucket := GetUserBucket(userID)
		if !bucket.Allow() {
			c.JSON(429, gin.H{
				"code": 429,
				"msg":  "Too many requests per user",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

