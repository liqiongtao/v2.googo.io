package gooredis

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// 连接地址
func WithAddr(addr string) FuncOption {
	return func(c *Config) {
		c.Addr = addr
	}
}

// 用户名
func WithUsername(username string) FuncOption {
	return func(c *Config) {
		c.Username = username
	}
}

// 密码
func WithPassword(password string) FuncOption {
	return func(c *Config) {
		c.Password = password
	}
}

// 数据库编号
func WithDB(db int) FuncOption {
	return func(c *Config) {
		c.DB = db
	}
}

// 前缀
func WithPrefix(prefix string) FuncOption {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// 连接池最大连接数
func WithPoolSize(poolSize int) FuncOption {
	return func(c *Config) {
		c.PoolSize = poolSize
	}
}

// 连接池最小空闲连接数
func WithMinIdleConns(minIdleConns int) FuncOption {
	return func(c *Config) {
		c.MinIdleConns = minIdleConns
	}
}

// 连接超时时间
func WithDialTimeout(dialTimeout time.Duration) FuncOption {
	return func(c *Config) {
		c.DialTimeout = dialTimeout
	}
}

// 读取超时时间
func WithReadTimeout(readTimeout time.Duration) FuncOption {
	return func(c *Config) {
		c.ReadTimeout = readTimeout
	}
}

// 写入超时时间
func WithWriteTimeout(writeTimeout time.Duration) FuncOption {
	return func(c *Config) {
		c.WriteTimeout = writeTimeout
	}
}

// 最大重试次数
func WithMaxRetries(maxRetries int) FuncOption {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// 重试间隔
func WithMinRetryBackoff(minRetryBackoff time.Duration) FuncOption {
	return func(c *Config) {
		c.MinRetryBackoff = minRetryBackoff
	}
}

// 最大重试间隔
func WithMaxRetryBackoff(maxRetryBackoff time.Duration) FuncOption {
	return func(c *Config) {
		c.MaxRetryBackoff = maxRetryBackoff
	}
}
