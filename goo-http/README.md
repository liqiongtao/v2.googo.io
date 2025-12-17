# goo-http

一个基于 [Gin](https://github.com/gin-gonic/gin) 的高性能 HTTP 框架封装库，提供了完整的中间件支持和统一的 API 响应格式。

## 功能特性

- 🔍 **Trace ID 追踪** - 自动生成和传递请求追踪 ID
- 📝 **日志记录** - 可扩展的日志接口，支持自定义日志实现
- 🌐 **CORS 支持** - 完整的跨域资源共享支持
- 🚦 **限流控制** - 基于令牌桶算法的限流器，支持多维度限流
- 🔐 **加密传输** - AES-256-GCM 加密，支持请求和响应加密
- 🎣 **响应钩子** - 灵活的响应处理钩子机制
- 📦 **统一响应** - 标准化的 API 响应格式
- ⚡ **性能优化** - Buffer 池复用，减少内存分配

## 安装

```bash
go get v2.googo.io/goo-http
```

## 快速开始

### 基础示例

```go
package main

import (
	"log"
	"v2.googo.io/goo-http"
)

func main() {
	// 创建服务器
	server := goohttp.New(
		goohttp.WithAddr(":8080"),
	)

	// 注册路由
	server.Get("/hello", func(ctx *goohttp.Context) {
		ctx.Success(map[string]string{
			"message": "Hello, World!",
		})
	})

	// 启动服务器
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
```

### 完整配置示例

```go
package main

import (
	"log"
	"v2.googo.io/goo-http"
)

func main() {
	// 创建限流器
	rateLimiter := goohttp.NewRateLimiter(&goohttp.RateLimitConfig{
		Rate:  100,              // 每秒 100 个请求
		Burst: 200,              // 突发 200 个请求
		KeyFunc: func(c *goohttp.Context) string {
			return c.ClientIP()  // 基于 IP 限流
		},
	})

	// 创建加密器
	key := make([]byte, 32)
	// 填充你的密钥...
	encryptor, _ := goohttp.NewAES256GCMEncryptor(key)

	// 创建服务器
	server := goohttp.New(
		goohttp.WithAddr(":8080"),
		goohttp.WithTraceIdHeader("X-Request-Id"),
		goohttp.WithEnableLog(true),
		goohttp.WithEnableCORS(true),
		goohttp.WithCORSConfig(&goohttp.CORSConfig{
			AllowOrigins: []string{"https://example.com"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		}),
		goohttp.WithEnableRateLimit(true),
		goohttp.WithRateLimit(rateLimiter),
		goohttp.WithEnableEncrypt(true),
		goohttp.WithEncryptor(encryptor),
	)

	// 注册路由
	server.Get("/api/users", func(ctx *goohttp.Context) {
		ctx.Success(map[string]interface{}{
			"users": []string{"user1", "user2"},
		})
	})

	// 启动服务器
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## API 文档

### Server

#### 创建服务器

```go
func New(opts ...ConfigOption) *Server
```

使用配置选项创建新的服务器实例。

#### 路由注册

```go
// HTTP 方法
server.Get(path string, handlers ...HandlerFunc)
server.Post(path string, handlers ...HandlerFunc)
server.Put(path string, handlers ...HandlerFunc)
server.Delete(path string, handlers ...HandlerFunc)
server.Patch(path string, handlers ...HandlerFunc)
server.Options(path string, handlers ...HandlerFunc)

// 路由组
group := server.Group("/api/v1")
group.Get("/users", handler)

// 静态文件
server.Static("/static", "./static")
server.StaticFile("/favicon.ico", "./favicon.ico")
```

#### 启动和关闭

```go
// 启动服务器（阻塞）
err := server.Run()

// 优雅关闭
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := server.Shutdown(ctx)
```

### Context

`Context` 是对 `gin.Context` 的封装，提供了便捷的方法。

#### 获取 Trace ID

```go
traceId := ctx.TraceId()
```

#### 获取客户端 IP

```go
ip := ctx.ClientIP()
```

#### 响应方法

```go
// 成功响应
ctx.Success(data any)
ctx.SuccessWithMessage(message string, data interface{})

// 错误响应
ctx.Error(code int, message string)
ctx.ErrorWithData(code int, message string, data interface{})
ctx.ErrorWithStatus(httpStatus int, code int, message string)

// 中止请求
ctx.Abort(httpStatus int, code int, message string)
```

### 配置选项

#### WithAddr

设置服务器监听地址。

```go
goohttp.WithAddr(":8080")
```

#### WithTraceIdHeader

设置 Trace ID 请求头名称。

```go
goohttp.WithTraceIdHeader("X-Request-Id")
```

#### WithEnableLog / WithLogger

启用日志并设置日志器。

```go
// 使用默认日志器
goohttp.WithEnableLog(true)

// 使用自定义日志器
type CustomLogger struct{}

func (l *CustomLogger) Info(ctx context.Context, msg string, fields ...interface{}) {
	// 实现日志逻辑
}

goohttp.WithLogger(&CustomLogger{})
```

#### WithEnableCORS / WithCORSConfig

启用 CORS 并配置。

```go
goohttp.WithEnableCORS(true)
goohttp.WithCORSConfig(&goohttp.CORSConfig{
	AllowOrigins:     []string{"https://example.com"},
	AllowMethods:     []string{"GET", "POST"},
	AllowHeaders:     []string{"Content-Type"},
	ExposeHeaders:    []string{"X-Trace-Id"},
	AllowCredentials: true,
	MaxAge:           86400,
})
```

#### WithEnableRateLimit / WithRateLimit

启用限流并配置限流器。

```go
rateLimiter := goohttp.NewRateLimiter(&goohttp.RateLimitConfig{
	Rate:  100,
	Burst: 200,
	KeyFunc: func(c *goohttp.Context) string {
		return c.ClientIP()
	},
	CleanupInterval: 5 * time.Minute,
	MaxIdleTime:     10 * time.Minute,
})

goohttp.WithEnableRateLimit(true)
goohttp.WithRateLimit(rateLimiter)
```

#### WithEnableEncrypt / WithEncryptor

启用加密并设置加密器。

```go
key := make([]byte, 32)
// 填充密钥
encryptor, _ := goohttp.NewAES256GCMEncryptor(key)

goohttp.WithEnableEncrypt(true)
goohttp.WithEncryptor(encryptor)
```

#### WithResponseHooks

设置响应钩子函数。

```go
hooks := []goohttp.ResponseHook{
	func(ctx *goohttp.Context, resp *goohttp.Response) {
		// 处理响应
		log.Printf("Response: %+v", resp)
	},
}

goohttp.WithResponseHooks(hooks)
```

## 中间件

### Trace 中间件

自动生成和传递 Trace ID。如果请求头中已存在 Trace ID，则使用现有的；否则生成新的 UUID。

### 日志中间件

记录请求和响应信息，包括：
- 请求方法、URI
- Trace ID
- 客户端 IP
- 响应状态码
- 请求耗时

### CORS 中间件

处理跨域请求，支持：
- 预检请求（OPTIONS）
- 自定义源、方法、请求头
- 凭证支持
- 缓存时间配置

### 限流中间件

基于令牌桶算法的限流，支持：
- 多限流器组合
- 自定义限流 key（如 IP、用户 ID 等）
- 自动清理不活跃的限流器

### 加密中间件

使用 AES-256-GCM 加密算法：
- 自动解密请求体
- 自动加密响应体
- 支持密钥动态更新

### 响应钩子中间件

在响应发送前执行自定义逻辑，可用于：
- 日志记录
- 指标收集
- 响应转换
- 审计追踪

## 响应格式

所有 API 响应遵循统一格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

- `code`: 业务状态码，0 表示成功
- `message`: 响应消息
- `data`: 响应数据
- `trace_id`: 追踪 ID

## 限流器

### 创建限流器

```go
rateLimiter := goohttp.NewRateLimiter(&goohttp.RateLimitConfig{
	Rate:  100,              // 每秒允许的请求数
	Burst: 200,              // 突发请求数
	KeyFunc: func(c *goohttp.Context) string {
		// 返回限流的 key，如 IP、用户 ID 等
		return c.ClientIP()
	},
	CleanupInterval: 5 * time.Minute,  // 清理间隔
	MaxIdleTime:     10 * time.Minute, // 最大空闲时间
})
```

### 更新限流配置

```go
rateLimiter.UpdateConfig(&goohttp.RateLimitConfig{
	Rate:  200,
	Burst: 400,
})
```

### 停止限流器

```go
rateLimiter.Stop()
```

## 加密器

### 创建加密器

```go
key := make([]byte, 32)
// 填充 32 字节密钥
copy(key, []byte("your-32-byte-secret-key-here!!"))

encryptor, err := goohttp.NewAES256GCMEncryptor(key)
if err != nil {
	log.Fatal(err)
}
```

### 动态更新密钥

```go
newKey := make([]byte, 32)
// 填充新密钥
err := encryptor.SetKey(newKey)
```

## 日志接口

实现 `Logger` 接口以使用自定义日志器：

```go
type Logger interface {
	Info(ctx context.Context, msg string, fields ...interface{})
	Error(ctx context.Context, msg string, fields ...interface{})
	Debug(ctx context.Context, msg string, fields ...interface{})
	Warn(ctx context.Context, msg string, fields ...interface{})
}
```

示例（集成 zap）：

```go
type ZapLogger struct {
	logger *zap.Logger
}

func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Info(msg, zap.Any("fields", fields))
}

// 使用
server := goohttp.New(
	goohttp.WithLogger(&ZapLogger{logger: zapLogger}),
)
```

## 中间件执行顺序

中间件按以下顺序执行：

1. **Trace 中间件** - 生成/获取 Trace ID
2. **日志中间件** - 记录请求信息
3. **CORS 中间件** - 处理跨域
4. **限流中间件** - 限流检查
5. **响应钩子中间件** - 捕获响应（在加密前）
6. **加密中间件** - 加解密处理（最后执行）

## 性能优化

- **Buffer 池**: 使用 `sync.Pool` 复用 buffer，减少内存分配
- **限流器清理**: 自动清理不活跃的限流器，防止内存泄漏
- **并发安全**: 所有共享资源都使用适当的锁保护

## 注意事项

1. **配置共享**: 多个 Server 实例会共享 `DefaultConfig`，建议为每个实例单独配置
2. **加密密钥**: 密钥必须为 32 字节，妥善保管密钥
3. **限流器清理**: 服务器关闭时会自动停止限流器的清理 goroutine
4. **响应格式**: 响应钩子只能处理 JSON 格式的响应

## 许可证

[添加你的许可证信息]

## 贡献

欢迎提交 Issue 和 Pull Request！
