package gooredis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// Config Redis 配置
type Config struct {
	// 连接地址，格式: host:port
	Addr string

	// 用户名（Redis 6.0+）
	Username string

	// 密码
	Password string

	// 数据库编号，默认 0
	DB int

	// 前缀
	Prefix string

	// 连接池最大连接数，默认 10
	PoolSize int

	// 连接池最小空闲连接数，默认 5
	MinIdleConns int

	// 连接超时时间，默认 5 秒
	DialTimeout time.Duration

	// 读取超时时间，默认 3 秒
	ReadTimeout time.Duration

	// 写入超时时间，默认 3 秒
	WriteTimeout time.Duration

	// 最大重试次数，默认 3
	MaxRetries int

	// 重试间隔，默认 128 毫秒
	MinRetryBackoff time.Duration

	// 最大重试间隔，默认 512 毫秒
	MaxRetryBackoff time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Addr:            "localhost:6379",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    5,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		MaxRetries:      3,
		MinRetryBackoff: 128 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toRedisOptions 转换为 go-redis 的 Options
func (c *Config) toRedisOptions() *redis.Options {
	opts := &redis.Options{
		Addr:            c.Addr,
		Username:        c.Username,
		Password:        c.Password,
		DB:              c.DB,
		PoolSize:        c.PoolSize,
		MinIdleConns:    c.MinIdleConns,
		DialTimeout:     c.DialTimeout,
		ReadTimeout:     c.ReadTimeout,
		WriteTimeout:    c.WriteTimeout,
		MaxRetries:      c.MaxRetries,
		MinRetryBackoff: c.MinRetryBackoff,
		MaxRetryBackoff: c.MaxRetryBackoff,
	}

	// 设置默认值
	if opts.PoolSize == 0 {
		opts.PoolSize = 10
	}
	if opts.MinIdleConns == 0 {
		opts.MinIdleConns = 5
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 3 * time.Second
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.MinRetryBackoff == 0 {
		opts.MinRetryBackoff = 128 * time.Millisecond
	}
	if opts.MaxRetryBackoff == 0 {
		opts.MaxRetryBackoff = 512 * time.Millisecond
	}

	return opts
}
