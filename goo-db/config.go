package goodb

import (
	"time"
)

// Config 数据库配置
type Config struct {
	// 数据库驱动: "mysql" 或 "postgres"
	Driver string

	// 数据源名称（连接字符串）
	DSN string

	// 连接池最大空闲连接数，默认 10
	MaxIdleConns int

	// 连接池最大打开连接数，默认 100
	MaxOpenConns int

	// 连接最大生存时间，默认 1 小时
	ConnMaxLifetime time.Duration

	// 连接最大空闲时间，默认 30 分钟
	ConnMaxIdleTime time.Duration

	// 是否显示 SQL 语句，默认 false
	ShowSQL bool

	// 日志级别，默认 0（不记录）
	LogLevel int

	// 慢查询阈值，默认 1 秒
	SlowQueryTime time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Driver:          "mysql",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		ShowSQL:         false,
		LogLevel:         0,
		SlowQueryTime:    1 * time.Second,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

