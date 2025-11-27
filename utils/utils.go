package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateOrderNo 生成订单号
func GenerateOrderNo() string {
	return fmt.Sprintf("ORD%d%s", time.Now().Unix(), uuid.New().String()[:8])
}

// IsSeckillTime 判断是否在秒杀时间内
func IsSeckillTime(startTime, endTime time.Time) bool {
	now := time.Now()
	return now.After(startTime) && now.Before(endTime)
}

// BeforeSeckillTime 判断是否在秒杀开始前
func BeforeSeckillTime(startTime time.Time) bool {
	return time.Now().Before(startTime)
}

// AfterSeckillTime 判断是否已过秒杀时间
func AfterSeckillTime(endTime time.Time) bool {
	return time.Now().After(endTime)
}

