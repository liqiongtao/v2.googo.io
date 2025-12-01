package goodb

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithDriver 设置数据库驱动
func WithDriver(driver string) FuncOption {
	return func(c *Config) {
		c.Driver = driver
	}
}

// WithDSN 设置数据源名称
func WithDSN(dsn string) FuncOption {
	return func(c *Config) {
		c.DSN = dsn
	}
}

// WithMaxIdleConns 设置连接池最大空闲连接数
func WithMaxIdleConns(maxIdleConns int) FuncOption {
	return func(c *Config) {
		c.MaxIdleConns = maxIdleConns
	}
}

// WithMaxOpenConns 设置连接池最大打开连接数
func WithMaxOpenConns(maxOpenConns int) FuncOption {
	return func(c *Config) {
		c.MaxOpenConns = maxOpenConns
	}
}

// WithConnMaxLifetime 设置连接最大生存时间
func WithConnMaxLifetime(connMaxLifetime time.Duration) FuncOption {
	return func(c *Config) {
		c.ConnMaxLifetime = connMaxLifetime
	}
}

// WithConnMaxIdleTime 设置连接最大空闲时间
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) FuncOption {
	return func(c *Config) {
		c.ConnMaxIdleTime = connMaxIdleTime
	}
}

// WithShowSQL 设置是否显示 SQL 语句
func WithShowSQL(showSQL bool) FuncOption {
	return func(c *Config) {
		c.ShowSQL = showSQL
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(logLevel int) FuncOption {
	return func(c *Config) {
		c.LogLevel = logLevel
	}
}

// WithSlowQueryTime 设置慢查询阈值
func WithSlowQueryTime(slowQueryTime time.Duration) FuncOption {
	return func(c *Config) {
		c.SlowQueryTime = slowQueryTime
	}
}

