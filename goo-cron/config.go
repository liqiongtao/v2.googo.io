package goocron

import (
	"time"

	cron "github.com/robfig/cron/v3"
)

// Config Cron 配置
type Config struct {
	// 时区，默认使用本地时区
	Location *time.Location

	// 解析器类型，默认使用标准解析器
	// 可选值: "standard" (标准解析器) 或 "second" (秒级解析器)
	Parser string

	// 是否记录任务执行日志，默认 false
	Logger bool

	// 链式选项，用于自定义 cron 行为
	Chain cron.Chain
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Location: time.Local,
		Parser:   "standard",
		Logger:   false,
		Chain:    cron.DefaultChain,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toCronOptions 转换为 robfig/cron 的 Options
func (c *Config) toCronOptions() []cron.Option {
	var opts []cron.Option

	// 设置时区
	if c.Location != nil {
		opts = append(opts, cron.WithLocation(c.Location))
	}

	// 设置解析器
	if c.Parser == "second" {
		opts = append(opts, cron.WithSeconds())
	} else {
		opts = append(opts, cron.WithParser(cron.NewParser(
			cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
		)))
	}

	// 设置链式选项
	if c.Chain != nil {
		opts = append(opts, cron.WithChain(c.Chain))
	}

	return opts
}
