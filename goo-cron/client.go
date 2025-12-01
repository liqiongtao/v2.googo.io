package goocron

import (
	"context"
	"sync"

	cron "github.com/robfig/cron/v3"
)

// Cron Cron 定时任务封装
type Cron struct {
	name    string
	config  *Config
	cron    *cron.Cron
	mu      sync.RWMutex
	entries map[cron.EntryID]string // 记录任务 ID 和名称的映射
}

// NewCron 创建新的 Cron 实例
func NewCron(name string, config *Config) (*Cron, error) {
	if config == nil {
		config = DefaultConfig()
	}

	opts := config.toCronOptions()
	c := cron.New(opts...)

	// 如果配置了日志，可以设置自定义日志记录器
	// 这里可以根据需要扩展日志功能

	instance := &Cron{
		name:    name,
		config:  config,
		cron:    c,
		entries: make(map[cron.EntryID]string),
	}

	return instance, nil
}

// Name 获取 Cron 实例名称
func (c *Cron) Name() string {
	return c.name
}

// Cron 获取底层的 cron.Cron 实例
func (c *Cron) Cron() *cron.Cron {
	return c.cron
}

// AddFunc 添加一个函数任务
// spec: cron 表达式，例如 "0 0 * * *" (每天午夜)
// cmd: 要执行的函数
// 返回任务 ID 和错误
func (c *Cron) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := c.cron.AddFunc(spec, cmd)
	if err != nil {
		return 0, err
	}

	c.entries[id] = spec
	return id, nil
}

// AddJob 添加一个 Job 任务
// spec: cron 表达式
// cmd: 要执行的 Job 接口实现
// 返回任务 ID 和错误
func (c *Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, err := c.cron.AddJob(spec, cmd)
	if err != nil {
		return 0, err
	}

	c.entries[id] = spec
	return id, nil
}

// Remove 移除一个任务
func (c *Cron) Remove(id cron.EntryID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cron.Remove(id)
	delete(c.entries, id)
}

// Entries 获取所有任务条目
func (c *Cron) Entries() []cron.Entry {
	return c.cron.Entries()
}

// Entry 获取指定 ID 的任务条目
func (c *Cron) Entry(id cron.EntryID) cron.Entry {
	return c.cron.Entry(id)
}

// Start 启动 Cron 调度器
func (c *Cron) Start() {
	c.cron.Start()
}

// Stop 停止 Cron 调度器（不会等待正在运行的任务完成）
func (c *Cron) Stop() context.Context {
	return c.cron.Stop()
}

// Run 运行 Cron 调度器（阻塞调用）
func (c *Cron) Run() {
	c.cron.Run()
}

// Close 关闭 Cron 实例
func (c *Cron) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx := c.cron.Stop()
	<-ctx.Done()
	c.entries = make(map[cron.EntryID]string)
	return nil
}
