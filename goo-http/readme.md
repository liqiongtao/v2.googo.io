# goo-http

基于 `github.com/gin-gonic/gin` 的 HTTP 服务端封装库，提供统一响应格式、请求追踪、加密传输、跨域处理等功能。

## 需求

1. 开发语言：golang
2. 包名: goohttp
3. 目录: goo-http
4. 功能需求:

   * 定义Server对象，基于 github.com/gin-gonic/gin
   * 定义Response对象
   * 定义Handler统一处理入口方法
   * 定义response hook 钩子
   * 定义日志处理，要有traceId进行追踪
   * 定义requestBody responseBody加密传输，加解密机制
   * 处理跨域问题
   * 处理异常访问问题
   * 定义包方法

## 功能特性

- ✅ **Server 对象**：基于 gin-gonic/gin 的服务器封装
- ✅ **统一响应格式**：标准化的 Response 结构，包含 code、message、data、traceId
- ✅ **Handler 统一处理**：统一的处理函数签名，自动处理错误和响应
- ✅ **响应钩子**：支持在响应返回前执行自定义逻辑
- ✅ **TraceId 追踪**：自动生成和传递 traceId，支持请求链路追踪
- ✅ **请求/响应加密**：支持 AES-256-GCM 加密，可自定义加密器
- ✅ **CORS 支持**：灵活的跨域配置
- ✅ **限流保护**：基于令牌桶算法的请求限流
- ✅ **日志记录**：自动记录请求和响应日志，支持 traceId 关联
- ✅ **多 Server 管理**：支持注册多个服务器实例
- ✅ **线程安全**：所有操作都经过互斥锁保护

## 快速开始

### 安装

```bash
go get v2.googo.io/goo-http
```

### 基本使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "v2.googo.io/goo-http"
)

func main() {
    // 创建配置
    config := goohttp.DefaultConfig(
        goohttp.WithAddress(":8080"),
        goohttp.WithMode("debug"),
        goohttp.WithEnableTrace(true),
    )
    
    // 注册默认服务器
    if err := goohttp.RegisterDefault(config); err != nil {
        panic(err)
    }
    
    // 获取服务器
    server, err := goohttp.Default()
    if err != nil {
        panic(err)
    }
    
    // 注册路由
    server.GET("/hello", func(c *gin.Context) (interface{}, error) {
        return map[string]string{
            "message": "Hello, World!",
        }, nil
    })
    
    // 启动服务器
    if err := server.Serve(); err != nil {
        panic(err)
    }
}
```

### 使用响应钩子

```go
// 添加响应钩子
config := goohttp.DefaultConfig(
    goohttp.WithResponseHook(func(c *gin.Context, resp *goohttp.Response) {
        // 在响应返回前执行自定义逻辑
        resp.Data = map[string]interface{}{
            "original": resp.Data,
            "timestamp": time.Now().Unix(),
        }
    }),
)

server, _ := goohttp.NewServer("my-server", config)
```

### 启用加密传输

```go
// 生成32字节密钥（AES-256需要32字节）
key := make([]byte, 32)
// ... 填充密钥数据 ...

config := goohttp.DefaultConfig(
    goohttp.WithEncryption(true, key),
)

server, _ := goohttp.NewServer("secure-server", config)
```

### 配置CORS

```go
config := goohttp.DefaultConfig(
    goohttp.WithCORS(
        true,
        []string{"https://example.com", "https://app.example.com"},
        []string{"GET", "POST", "PUT", "DELETE"},
        []string{"Content-Type", "Authorization"},
    ),
    goohttp.WithCORSAllowCredentials(true),
)
```

### 启用限流

```go
config := goohttp.DefaultConfig(
    goohttp.WithRateLimit(true, 100, 200), // 每秒100个请求，突发200个
)
```

## API 文档

### Server

HTTP 服务器对象。

**方法：**
- `NewServer(name string, config *Config) (*Server, error)` - 创建新服务器
- `Name() string` - 获取服务器名称
- `Engine() *gin.Engine` - 获取 Gin 引擎
- `GET(relativePath string, handler HandlerFunc)` - 注册 GET 路由
- `POST(relativePath string, handler HandlerFunc)` - 注册 POST 路由
- `PUT(relativePath string, handler HandlerFunc)` - 注册 PUT 路由
- `DELETE(relativePath string, handler HandlerFunc)` - 注册 DELETE 路由
- `PATCH(relativePath string, handler HandlerFunc)` - 注册 PATCH 路由
- `OPTIONS(relativePath string, handler HandlerFunc)` - 注册 OPTIONS 路由
- `Any(relativePath string, handler HandlerFunc)` - 注册任意方法的路由
- `Use(middleware ...gin.HandlerFunc)` - 添加中间件
- `Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup` - 创建路由组
- `Serve() error` - 启动服务器（阻塞调用）
- `ServeTLS(certFile, keyFile string) error` - 启动 HTTPS 服务器
- `Shutdown() error` - 优雅关闭服务器
- `Close() error` - 关闭服务器

### Response

统一响应结构。

```go
type Response struct {
    Code    int         `json:"code"`    // 业务状态码
    Message string      `json:"message"` // 响应消息
    Data    interface{} `json:"data"`    // 响应数据
    TraceId string      `json:"trace_id,omitempty"` // 追踪ID
}
```

**方法：**
- `Success(c *gin.Context, data interface{})` - 成功响应
- `SuccessWithMessage(c *gin.Context, message string, data interface{})` - 带消息的成功响应
- `Error(c *gin.Context, code int, message string)` - 错误响应
- `ErrorWithData(c *gin.Context, code int, message string, data interface{})` - 带数据的错误响应
- `BadRequest(c *gin.Context, message string)` - 400错误响应
- `Unauthorized(c *gin.Context, message string)` - 401错误响应
- `Forbidden(c *gin.Context, message string)` - 403错误响应
- `NotFound(c *gin.Context, message string)` - 404错误响应
- `InternalServerError(c *gin.Context, message string)` - 500错误响应

### Handler

统一处理函数类型。

```go
type HandlerFunc func(*gin.Context) (interface{}, error)
```

**方法：**
- `Handler(handler HandlerFunc) gin.HandlerFunc` - 将 HandlerFunc 转换为 gin.HandlerFunc

### Config

服务器配置对象。

**配置选项：**
- `WithAddress(address string) FuncOption` - 设置服务器地址
- `WithMode(mode string) FuncOption` - 设置运行模式（debug/release/test）
- `WithEnableTrace(enable bool) FuncOption` - 设置是否启用追踪
- `WithEncryption(enable bool, key []byte) FuncOption` - 设置加密配置
- `WithEncryptor(encryptor Encryptor, decryptor Decryptor) FuncOption` - 设置自定义加密器
- `WithCORS(enable bool, origins []string, methods []string, headers []string) FuncOption` - 设置CORS配置
- `WithCORSAllowCredentials(allow bool) FuncOption` - 设置是否允许携带凭证
- `WithCORSMaxAge(maxAge time.Duration) FuncOption` - 设置预检请求缓存时间
- `WithRateLimit(enable bool, rate float64, burst int) FuncOption` - 设置限流配置
- `WithRateLimitKeyFunc(keyFunc func(*gin.Context) string) FuncOption` - 设置限流键生成函数
- `WithLog(enable bool, logRequestBody bool, logResponseBody bool) FuncOption` - 设置日志配置
- `WithLogMaxBodySize(maxRequestBodySize, maxResponseBodySize int) FuncOption` - 设置日志记录的最大请求体和响应体大小
- `WithResponseHook(hook ResponseHook) FuncOption` - 添加响应钩子
- `WithGinOptions(opts ...gin.Option) FuncOption` - 设置Gin引擎选项

### 包方法

- `Register(name string, config *Config) error` - 注册一个服务器
- `RegisterDefault(config *Config) error` - 注册默认服务器
- `Unregister(name string) error` - 注销一个服务器
- `UnregisterDefault() error` - 注销默认服务器
- `GetServer(name string) (*Server, error)` - 获取指定名称的服务器
- `Default() (*Server, error)` - 获取默认服务器
- `CloseAll() error` - 关闭所有服务器

## 使用建议

1. **TraceId 追踪**：建议在生产环境中启用 traceId，便于问题排查和链路追踪
2. **加密传输**：如果使用加密传输，确保客户端和服务端使用相同的密钥和算法
3. **CORS 配置**：生产环境应该明确指定允许的源，避免使用 `*`
4. **限流设置**：根据实际业务需求合理设置限流参数，避免误杀正常请求
5. **日志记录**：生产环境建议关闭请求体和响应体的详细日志，避免日志过大
6. **错误处理**：Handler 函数返回的错误会自动转换为 500 错误响应，可以根据需要自定义错误处理逻辑
7. **响应钩子**：响应钩子会在响应返回前执行，可以用于添加通用字段或进行数据转换

## 注意事项

1. **加密密钥**：AES-256-GCM 需要 32 字节的密钥，请妥善保管密钥，不要硬编码在代码中
2. **限流实现**：当前限流实现是全局限流，如果需要按 IP 或其他维度限流，需要自定义 `KeyFunc`
3. **响应体加密**：响应体加密会在响应写入后执行，可能会影响性能，建议仅在必要时启用
4. **请求体解密**：如果启用加密，客户端需要在请求头中标识加密内容，或者使用约定的格式
5. **TraceId 传递**：TraceId 会自动从请求头 `X-Trace-Id` 获取，如果没有则自动生成
6. **优雅关闭**：使用 `Shutdown()` 方法可以优雅关闭服务器，等待正在处理的请求完成
7. **多服务器管理**：支持注册多个服务器实例，通过名称区分，适合微服务场景
8. **线程安全**：所有全局操作都是线程安全的，可以在多个 goroutine 中安全使用
