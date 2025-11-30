# goo-log 日志库

一个功能完整的 Go 语言日志库，支持多种输出适配器、日志级别、文件切割和压缩等功能。

## 功能特性

1. **日志级别定义**: DEBUG INFO WARN ERROR PANIC FATAL
2. **日志级别对应颜色定义**: blue green yellow red white magenta
3. **链式调用**: 支持流畅的链式调用 API
4. **多适配器输出**: 支持 console, file, es, kafka 等适配器
5. **日志格式**: 支持 JSON 和控制台文本格式
6. **文件管理**: 支持自定义目录、文件名、文件大小
7. **自动切割**: 文件超过设置大小后自动切割
8. **自动压缩**: 支持设置日志文件保留天数，自动进行 gzip 压缩

## 快速开始

### 基本使用

```go
package main

import (
    "v2.googo.io/goo-log"
    "v2.googo.io/goo-log/adapters"
)

func main() {
    // 使用控制台适配器
    consoleAdapter := adapters.NewConsoleAdapter(false)
    goolog.SetAdapter(consoleAdapter)
    goolog.SetLevel(goolog.INFO)

    // 基本日志
    goolog.Info("这是一条信息日志")
    goolog.Warn("这是一条警告日志")
    goolog.Error("这是一条错误日志")
}
```

### 使用标签和字段

```go
// 使用标签
goolog.WithTag("user", "admin").Info("用户登录")

// 使用字段
goolog.WithField("userId", 12345).
    WithField("action", "login").
    Info("用户操作")

// 使用格式化字段
goolog.WithFieldF("price", "%.2f", 99.99).Info("价格信息")
```

### 使用上下文

```go
import "v2.googo.io/goo-context"

// 链式调用创建上下文
ctx := goocontext.WithAppName(nil, "my-app").WithTraceId()
goolog.WithContext(ctx).Info("带上下文的日志")

// 或者使用 Default 创建后链式调用
ctx = goocontext.Default(nil).WithAppName("my-app").WithTraceId()
goolog.WithContext(ctx).Info("带上下文的日志")
```

### 使用追踪信息

```go
// 强制添加追踪信息
goolog.WithTrace().Error("错误信息，包含堆栈追踪")

// 设置追踪级别（达到此级别及以上自动添加追踪）
goolog.SetTraceLevel(goolog.ERROR)
```

### 文件适配器

```go
fileAdapter, err := adapters.NewFileAdapter(adapters.FileConfig{
    Dir:        "logs",                    // 日志目录
    FileName:   "2006-01-02.log",         // 文件名模板（支持日期格式，会格式化为当前日期，如：2024-01-15.log）
    MaxSize:    200 * 1024 * 1024,        // 最大文件大小（200MB）
    RetainDays: 30,                        // 保留天数，默认 30
    UseJSON:    false,                     // 是否使用 JSON 格式
})
if err != nil {
    panic(err)
}
defer fileAdapter.Close()

goolog.SetAdapter(fileAdapter)
```

### ES 适配器

```go
esAdapter := adapters.NewESAdapter(adapters.ESConfig{
    URL:      "http://localhost:9200",
    Index:    "goolog",
    UseAsync: true,
})
goolog.SetAdapter(esAdapter)
```

### Kafka 适配器

```go
// 注意：需要提供实际的 Kafka producer
kafkaAdapter := adapters.NewKafkaAdapter(adapters.KafkaConfig{
    Topic:    "logs",
    Producer: yourKafkaProducer,
})
goolog.SetAdapter(kafkaAdapter)
```

### 链式调用

```go
goolog.WithTag("api", "v1").
    WithField("method", "GET").
    WithField("path", "/api/users").
    WithContext(ctx).
    Info("API 请求")
```

### 添加钩子函数

```go
goolog.AddHook(func(msg *goolog.Message) {
    // 可以在这里做一些额外处理，比如发送到监控系统
    // 注意：钩子函数在独立的 goroutine 中执行
})
```

## API 文档

### 日志级别

- `DEBUG`: 调试信息
- `INFO`: 一般信息
- `WARN`: 警告信息
- `ERROR`: 错误信息
- `PANIC`: 恐慌信息（会触发 panic）
- `FATAL`: 致命错误（会调用 os.Exit(1)）

### 全局函数

- `SetLevel(level Level)`: 设置日志级别
- `SetTraceLevel(level Level)`: 设置追踪级别
- `SetAdapter(adapter Adapter)`: 设置适配器
- `AddHook(fn func(msg *Message))`: 添加钩子函数
- `WithTag(tags ...any)`: 创建带标签的 Entry
- `WithField(field string, value any)`: 创建带字段的 Entry
- `WithFieldF(field string, format string, args ...any)`: 创建带格式化字段的 Entry
- `WithContext(ctx *goocontext.Context)`: 从上下文创建 Entry
- `WithTrace()`: 创建带追踪信息的 Entry
- `Debug/Info/Warn/Error/Panic/Fatal(v ...any)`: 记录日志
- `DebugF/InfoF/WarnF/ErrorF/PanicF/FatalF(format string, v ...any)`: 格式化记录日志

### 文件适配器配置

- `Dir`: 日志目录，默认 "logs"
- `FileName`: 文件名模板，支持 Go 时间格式，默认 "2006-01-02.log"（会格式化为当前日期，如：2024-01-15.log）
- `MaxSize`: 最大文件大小（字节），默认 200MB
- `RetainDays`: 保留天数，默认 30 天
- `UseJSON`: 是否使用 JSON 格式，默认 false

### 文件切割规则

- 当日志文件超过 `MaxSize` 时，会自动切割为 `文件名.1.log`、`文件名.2.log` 等
- 每天会自动创建新的日志文件
- 超过保留天数的日志文件会自动进行 gzip 压缩
- 压缩后的文件格式为 `文件名.log.gz`

## 注意事项

1. 文件适配器会在后台自动进行压缩和清理，使用完毕后建议调用 `Close()` 方法
2. ES 和 Kafka 适配器是基础实现，实际使用时可能需要根据具体的客户端库进行调整
3. 钩子函数在独立的 goroutine 中执行，需要注意线程安全
4. Entry 对象会被复用，使用对象池提高性能
