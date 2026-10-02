package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 配置
type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Client 全局 Redis 客户端
var Client *redis.Client
var ctx = context.Background()

// discardLogger 屏蔽 go-redis 内部的连接池重试日志（连不上时由 Init 统一告警）
type discardLogger struct{}

func (discardLogger) Printf(_ context.Context, _ string, _ ...interface{}) {}

// Init 初始化 Redis
func Init(cfg *Config) (*redis.Client, error) {
	redis.SetLogger(discardLogger{})
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := Client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return Client, nil
}

// --- 快捷方法 ---

// Set 设置键值
func Set(key string, value interface{}, ttl time.Duration) error {
	return Client.Set(ctx, key, value, ttl).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Del 删除键
func Del(keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	n, err := Client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetJSON 设置 JSON 对象（序列化后存储）
func SetJSON(key string, value interface{}, ttl time.Duration) error {
	return Client.Set(ctx, key, value, ttl).Err()
}

// GetJSON 获取 JSON 对象
func GetJSON(key string, dest interface{}) error {
	return Client.Get(ctx, key).Scan(dest)
}
