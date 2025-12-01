# goo-cron 定时任务

## 需求

1. 开发语言：golang
2. 包名: goocron
3. 目录: goo-cron
4. 功能需求:

   * 定义Cron对象，基于 github.com/robfig/cron/v3
   * 定义Config对象
   * 定义包方法

## 功能特性

- ✅ 基于 `github.com/robfig/cron/v3` 封装
- ✅ 支持多 Cron 实例切换（通过名称管理多个定时任务调度器）
- ✅ 支持标准 cron 表达式和秒级 cron 表达式
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局 Cron 实例管理
- ✅ 支持时区配置
- ✅ 支持链式选项（Chain）自定义任务行为
- ✅ 支持添加函数任务和 Job 任务

## 快速开始

### 安装依赖

```bash
go get github.com/robfig/cron/v3
```

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    
    "v2.googo.io/goo-cron"
)

func main() {
    // 1. 注册默认 Cron 实例
    config := goocron.DefaultConfig()
    if err := goocron.RegisterDefault(config); err != nil {
        panic(err)
    }
    
    // 2. 添加定时任务（每天午夜执行）
    id, err := goocron.AddFunc("0 0 * * *", func() {
        fmt.Println("执行定时任务:", time.Now())
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("任务 ID:", id)
    
    // 3. 启动 Cron 调度器
    goocron.Start()
    
    // 4. 保持程序运行
    select {}
}
```

### 使用 Job 接口

```go
type MyJob struct {
    Name string
}

func (j MyJob) Run() {
    fmt.Printf("执行任务: %s, 时间: %s\n", j.Name, time.Now())
}

func main() {
    config := goocron.DefaultConfig()
    goocron.RegisterDefault(config)
    
    // 添加 Job 任务
    job := MyJob{Name: "清理任务"}
    goocron.AddJob("0 */1 * * *", job) // 每小时执行一次
    
    goocron.Start()
    select {}
}
```

### 多 Cron 实例切换

```go
// 注册多个 Cron 实例
goocron.Register("task1", &goocron.Config{
    Location: time.UTC,
    Parser:   "standard",
})

goocron.Register("task2", &goocron.Config{
    Location: time.FixedZone("CST", 8*3600),
    Parser:   "second", // 使用秒级解析器
})

// 切换使用不同的 Cron 实例
cron1, _ := goocron.GetCron("task1")
cron1.AddFunc("0 0 * * *", func() {
    fmt.Println("任务1执行")
})

cron2, _ := goocron.GetCron("task2")
cron2.AddFunc("*/5 * * * * *", func() { // 每5秒执行
    fmt.Println("任务2执行")
})

cron1.Start()
cron2.Start()
```

### 使用配置选项

```go
// 使用函数式选项配置
config := goocron.DefaultConfig(
    goocron.WithLocation(time.UTC),
    goocron.WithParser("second"),
    goocron.WithLogger(true),
)

goocron.RegisterDefault(config)
```

### Cron 表达式说明

- **标准格式**: `分 时 日 月 星期`
  - `0 0 * * *` - 每天午夜
  - `0 */1 * * *` - 每小时
  - `0 9 * * 1-5` - 工作日上午9点

- **秒级格式**: `秒 分 时 日 月 星期`
  - `*/5 * * * * *` - 每5秒
  - `0 */1 * * * *` - 每分钟

## API 文档

### Config 配置对象

```go
type Config struct {
    Location *time.Location  // 时区，默认使用本地时区
    Parser   string          // 解析器类型: "standard" 或 "second"
    Logger   bool            // 是否记录任务执行日志，默认 false
    Chain    cron.Chain      // 链式选项，用于自定义 cron 行为
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### Cron 对象

```go
type Cron struct {
    // ...
}

// NewCron 创建新的 Cron 实例
func NewCron(name string, config *Config) (*Cron, error)

// Name 获取 Cron 实例名称
func (c *Cron) Name() string

// Cron 获取底层的 cron.Cron 实例
func (c *Cron) Cron() *cron.Cron

// AddFunc 添加一个函数任务
func (c *Cron) AddFunc(spec string, cmd func()) (cron.EntryID, error)

// AddJob 添加一个 Job 任务
func (c *Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error)

// Remove 移除一个任务
func (c *Cron) Remove(id cron.EntryID)

// Entries 获取所有任务条目
func (c *Cron) Entries() []cron.Entry

// Entry 获取指定 ID 的任务条目
func (c *Cron) Entry(id cron.EntryID) cron.Entry

// Start 启动 Cron 调度器
func (c *Cron) Start()

// Stop 停止 Cron 调度器（不会等待正在运行的任务完成）
func (c *Cron) Stop() context.Context

// Run 运行 Cron 调度器（阻塞调用）
func (c *Cron) Run()

// Close 关闭 Cron 实例
func (c *Cron) Close() error
```

### 包级别方法

```go
// Register 注册一个 Cron 实例（支持多 Cron 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认 Cron 实例
func RegisterDefault(config *Config) error

// Unregister 注销一个 Cron 实例
func Unregister(name string) error

// UnregisterDefault 注销默认 Cron 实例
func UnregisterDefault() error

// GetCron 获取指定名称的 Cron 实例
func GetCron(name string) (*Cron, error)

// Default 获取默认 Cron 实例
func Default() (*Cron, error)

// AddFunc 向默认 Cron 实例添加函数任务
func AddFunc(spec string, cmd func()) (cron.EntryID, error)

// AddJob 向默认 Cron 实例添加 Job 任务
func AddJob(spec string, cmd cron.Job) (cron.EntryID, error)

// Remove 从默认 Cron 实例移除任务
func Remove(id cron.EntryID) error

// Start 启动默认 Cron 实例
func Start() error

// Stop 停止默认 Cron 实例
func Stop() error

// Run 运行默认 Cron 实例（阻塞调用）
func Run() error

// CloseAll 关闭所有 Cron 实例
func CloseAll() error
```

### 配置选项函数

```go
// WithLocation 设置时区
func WithLocation(location *time.Location) FuncOption

// WithParser 设置解析器类型
func WithParser(parser string) FuncOption

// WithLogger 设置是否记录任务执行日志
func WithLogger(logger bool) FuncOption

// WithChain 设置链式选项
func WithChain(chain cron.Chain) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 Cron 实例
2. **资源管理**: 使用 `CloseAll()` 在应用退出时关闭所有 Cron 实例
3. **多实例场景**: 当需要不同的时区或解析器配置时，使用不同的名称注册多个 Cron 实例
4. **任务管理**: 保存任务 ID，以便后续可以移除或查询任务
5. **错误处理**: 始终检查 `Register()`、`AddFunc()`、`AddJob()` 等方法的返回值
6. **阻塞 vs 非阻塞**: 使用 `Start()` 非阻塞启动，使用 `Run()` 阻塞运行（适合主程序）
7. **任务执行时间**: 注意任务的执行时间，避免任务执行时间过长影响后续任务

## 注意事项

1. **依赖版本**: 本库基于 `github.com/robfig/cron/v3`，确保已正确安装依赖
2. **线程安全**: 全局 Cron 实例管理是线程安全的，可以并发使用
3. **任务执行**: 默认情况下，如果前一个任务还在执行，下一个任务会等待。如需并发执行，可以使用 `cron.WithChain(cron.SkipIfStillRunning(...))`
4. **资源释放**: 使用完毕后记得调用 `Close()` 或 `CloseAll()` 释放资源
5. **默认实例**: 如果没有设置默认 Cron 实例，使用 `Default()` 等方法会返回错误
6. **Cron 表达式**: 确保 cron 表达式格式正确，标准格式为 5 个字段，秒级格式为 6 个字段
7. **时区设置**: 注意时区配置，确保任务在预期的时间执行
