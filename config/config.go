package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Seckill  SeckillConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type SeckillConfig struct {
	TokenPrefix      string
	StockPrefix      string
	OrderPrefix      string
	LockPrefix       string
	TokenExpire      int
	PreheatKey       string
	MaxConcurrency   int
	RateLimitPerUser int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("SERVER_MODE", "debug"),
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "3306"),
			User:         getEnv("DB_USER", "root"),
			Password:     getEnv("DB_PASSWORD", "password"),
			Database:     getEnv("DB_NAME", "seckill"),
			MaxOpenConns: 100,
			MaxIdleConns: 10,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
			PoolSize: 100,
		},
		Seckill: SeckillConfig{
			TokenPrefix:      "seckill:token:",
			StockPrefix:      "seckill:stock:",
			OrderPrefix:      "seckill:order:",
			LockPrefix:       "seckill:lock:",
			TokenExpire:      3600,
			PreheatKey:       "seckill:preheat:",
			MaxConcurrency:   10000,
			RateLimitPerUser: 5,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

