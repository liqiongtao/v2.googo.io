package goocron

import (
	"sync"

	cron "github.com/robfig/cron/v3"
)

var (
	crons       = make(map[string]*Cron)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个 Cron 实例（支持多 Cron 切换）
func Register(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	cron, err := NewCron(name, config)
	if err != nil {
		return err
	}

	crons[name] = cron
	return nil
}

// RegisterDefault 注册默认 Cron 实例
func RegisterDefault(config *Config) error {
	return Register("default", config)
}

// Unregister 注销一个 Cron 实例
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()

	cron, ok := crons[name]
	if !ok {
		return nil
	}

	delete(crons, name)
	return cron.Close()
}

// UnregisterDefault 注销默认 Cron 实例
func UnregisterDefault() error {
	return Unregister("default")
}

// GetCron 获取指定名称的 Cron 实例
func GetCron(name string) (*Cron, error) {
	mu.RLock()
	defer mu.RUnlock()

	cron, ok := crons[name]
	if !ok {
		return nil, ErrCronNotFound
	}

	return cron, nil
}

// Default 获取默认 Cron 实例
func Default() (*Cron, error) {
	return GetCron(defaultName)
}

// AddFunc 向默认 Cron 实例添加函数任务
func AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	c, err := Default()
	if err != nil {
		return 0, err
	}
	return c.AddFunc(spec, cmd)
}

// AddJob 向默认 Cron 实例添加 Job 任务
func AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	c, err := Default()
	if err != nil {
		return 0, err
	}
	return c.AddJob(spec, cmd)
}

// Remove 从默认 Cron 实例移除任务
func Remove(id cron.EntryID) error {
	c, err := Default()
	if err != nil {
		return err
	}
	c.Remove(id)
	return nil
}

// Start 启动默认 Cron 实例
func Start() error {
	c, err := Default()
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

// Stop 停止默认 Cron 实例
func Stop() error {
	c, err := Default()
	if err != nil {
		return err
	}
	c.Stop()
	return nil
}

// Run 运行默认 Cron 实例（阻塞调用）
func Run() error {
	c, err := Default()
	if err != nil {
		return err
	}
	c.Run()
	return nil
}

// CloseAll 关闭所有 Cron 实例
func CloseAll() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	for name, cron := range crons {
		if closeErr := cron.Close(); closeErr != nil {
			err = closeErr
		}
		delete(crons, name)
	}

	return err
}
