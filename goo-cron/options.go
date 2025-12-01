package goocron

import (
	"time"

	cron "github.com/robfig/cron/v3"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithLocation 设置时区
func WithLocation(location *time.Location) FuncOption {
	return func(c *Config) {
		c.Location = location
	}
}

// WithParser 设置解析器类型
// parser: "standard" (标准解析器) 或 "second" (秒级解析器)
func WithParser(parser string) FuncOption {
	return func(c *Config) {
		c.Parser = parser
	}
}

// WithLogger 设置是否记录任务执行日志
func WithLogger(logger bool) FuncOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithChain 设置链式选项
func WithChain(chain cron.Chain) FuncOption {
	return func(c *Config) {
		c.Chain = chain
	}
}
