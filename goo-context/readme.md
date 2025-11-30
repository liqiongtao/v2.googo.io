# goo-context 上下文库

一个功能完整的 Go 语言全局上下文库，封装了 `context.Context`，提供增强功能，支持应用名称、追踪ID、类型自动转换等功能。

## 功能特性

1. **集成 context.Context**: 完全兼容标准库的 `context.Context` 接口
2. **键值对存储**: 支持设置和获取任意类型的键值对
3. **应用标识**: 支持设置和获取全局 AppName 和 TraceId
4. **上下文控制**: 支持取消、超时、截止时间等控制
5. **信号处理**: 支持监听系统信号并自动取消上下文
6. **框架集成**: 提供 Gin 和 gRPC 框架的集成支持
7. **类型自动转换**: 支持多种类型之间的自动转换（string, int, int32, int64, float32, float64, bool）

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "v2.googo.io/goo-context"
)

func main() {
    // 创建默认上下文并链式设置键值对
    ctx := goocontext.Default(context.Background()).
        WithValue("user-id", "12345").
        WithValue("username", "testuser")
    
    // 使用方法获取值
    userId := ctx.ValueString("user-id")
    username := ctx.ValueString("username")
}
```

### 设置应用名称和追踪ID

```go
// 使用 Context 方法（支持链式调用）
ctx := goocontext.Default(context.Background()).
    WithAppName("my-app").
    WithTraceId()  // 自动生成UUID

// 设置自定义追踪ID
ctx = ctx.WithTraceId("custom-trace-id-123")

// 获取应用名称和追踪ID
appName := ctx.AppName()
traceId := ctx.TraceId()
```

### 上下文控制

```go
// 创建可取消的上下文
ctx, cancel := ctx.WithCancel()
defer cancel()

// 创建带超时的上下文（5秒）
ctx, cancel := ctx.WithTimeout(5 * time.Second)
defer cancel()

// 创建带截止时间的上下文
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := ctx.WithDeadline(deadline)
defer cancel()

// 监听系统信号（默认监听 SIGUSR1, SIGUSR2, SIGHUP, SIGTERM, SIGQUIT, SIGINT）
ctx = ctx.WithSignalNotify()

// 监听指定信号
ctx = ctx.WithSignalNotify(syscall.SIGTERM, syscall.SIGINT)
```

### 类型自动转换

```go
// 链式设置不同类型的值
ctx := goocontext.Default(context.Background()).
    WithValue("str_val", "123").
    WithValue("int_val", 456).
    WithValue("float_val", 3.14).
    WithValue("bool_val", true)

// ValueInt64 支持从 string, int, int32, float32, float64, bool 自动转换
int64Val := ctx.ValueInt64("str_val")   // "123" -> 123
int64Val = ctx.ValueInt64("int_val")    // 456 -> 456
int64Val = ctx.ValueInt64("float_val")  // 3.14 -> 3
int64Val = ctx.ValueInt64("bool_val")   // true -> 1

// ValueString 支持从 int, int32, int64, float32, float64, bool 自动转换
strVal := ctx.ValueString("int_val")    // 456 -> "456"
strVal = ctx.ValueString("float_val")   // 3.14 -> "3.14"
strVal = ctx.ValueString("bool_val")    // true -> "true"

// ValueBool 支持从 string, int, int32, int64, float32, float64 自动转换
// 对于 string: "", "0", "nil", "null", "false", "no", "off" 返回 false，其他返回 true
boolVal := ctx.ValueBool("str_val")     // "123" -> true
boolVal = ctx.ValueBool("int_val")      // 456 -> true
boolVal = ctx.ValueBool("zero_val")     // 0 -> false
```

### Gin 框架集成

```go
import (
    "github.com/gin-gonic/gin"
    "v2.googo.io/goo-context"
)

func handler(c *gin.Context) {
    // 从 gin.Context 创建上下文
    ctx := goocontext.FromGinContext(c)
    
    // 或者从现有上下文更新
    parentCtx := goocontext.Default(context.Background()).
        WithAppName("my-app").
        WithTraceId()
    ctx = parentCtx.WithGinContext(c)
    
    // 使用上下文
    appName := ctx.AppName()
    traceId := ctx.TraceId()
}
```

### gRPC 框架集成

```go
import (
    "context"
    "v2.googo.io/goo-context"
    "google.golang.org/grpc"
)

// 服务端：从 gRPC context 提取上下文信息
func (s *Server) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    gooCtx := goocontext.FromGrpcContext(ctx)
    appName := gooCtx.AppName()
    traceId := gooCtx.TraceId()
    // ...
}

// 客户端：将上下文信息添加到 gRPC metadata
func callGrpc(ctx context.Context) {
    // 使用链式调用创建上下文
    gooCtx := goocontext.Default(context.Background()).
        WithAppName("my-app").
        WithTraceId()
    
    // 添加额外的 metadata
    gooCtx = gooCtx.WithGrpcContext("custom-key", "custom-value")
    
    // 使用 gooCtx.Context 作为 gRPC 调用的 context
    resp, err := client.MyMethod(gooCtx.Context, req)
}
```

### 获取不同类型的值

```go
// 链式设置不同类型的值
ctx := goocontext.Default(context.Background()).
    WithValue("str", "hello").
    WithValue("int", 42).
    WithValue("int32", int32(100)).
    WithValue("int64", int64(200)).
    WithValue("float32", float32(3.14)).
    WithValue("float64", 2.718).
    WithValue("bool", true)

// 使用类型安全的获取方法（支持自动转换）
str := ctx.ValueString("str")
intVal := ctx.ValueInt("int")
int32Val := ctx.ValueInt32("int32")
int64Val := ctx.ValueInt64("int64")
float32Val := ctx.ValueFloat32("float32")
float64Val := ctx.ValueFloat64("float64")
boolVal := ctx.ValueBool("bool")

// 获取原始值（需要手动类型断言）
anyVal := ctx.ValueAny("str")
```

## API 文档

### Context 结构体

`Context` 封装了 `context.Context`，提供增强功能：

```go
type Context struct {
    context.Context
}
```

### Context 方法

#### WithValue
设置键值对到上下文中：
```go
func (c *Context) WithValue(key string, value any) *Context
```

#### WithAppName
在当前上下文上设置应用名称：
```go
func (c *Context) WithAppName(appName string, args ...any) *Context
```
- 支持格式化字符串（类似 `fmt.Sprintf`）
- 支持链式调用

#### WithTraceId
在当前上下文上设置或生成追踪ID：
```go
func (c *Context) WithTraceId(traceId ...string) *Context
```
- 如果不提供 `traceId`，会自动生成 UUID
- 支持链式调用

#### AppName
获取应用名称：
```go
func (c *Context) AppName() string
```
- 支持多种 key 格式：`AppName`, `app-name`, `app_name`

#### TraceId
获取追踪ID：
```go
func (c *Context) TraceId() string
```
- 支持多种 key 格式：`TraceId`, `trace-id`, `trace_id`, `request_id`

#### WithCancel
创建一个可取消的上下文：
```go
func (c *Context) WithCancel() (*Context, context.CancelFunc)
```

#### WithTimeout
创建一个带超时的上下文：
```go
func (c *Context) WithTimeout(d time.Duration) (*Context, context.CancelFunc)
```

#### WithDeadline
创建一个带截止时间的上下文：
```go
func (c *Context) WithDeadline(d time.Time) (*Context, context.CancelFunc)
```

#### WithSignalNotify
创建一个监听系统信号的上下文：
```go
func (c *Context) WithSignalNotify(signals ...os.Signal) *Context
```
- 如果不指定 `signals`，默认监听：SIGUSR1, SIGUSR2, SIGHUP, SIGTERM, SIGQUIT, SIGINT
- 当接收到信号时，上下文会被自动取消

#### WithGinContext
从 gin.Context 创建或更新上下文：
```go
func (c *Context) WithGinContext(ginCtx *gin.Context) *Context
```
- 自动从 gin.Context 中提取 `app-name` 和 `trace-id`
- 如果 gin.Context 中没有这些值，会从当前上下文继承或自动生成

#### WithGrpcContext
将上下文信息添加到 gRPC metadata：
```go
func (c *Context) WithGrpcContext(kvs ...string) *Context
```
- 自动添加 `app-name` 和 `trace-id` 到 metadata
- 可以额外指定其他 key-value 对（必须是偶数个参数）

### 包级别函数

#### Default
创建默认上下文：
```go
func Default(parent context.Context) *Context
```

#### WithAppName
设置应用名称（包级别函数，用于从标准库 context 创建）：
```go
func WithAppName(parent context.Context, appName string, args ...any) *Context
```
- 支持格式化字符串（类似 `fmt.Sprintf`）
- 用于从标准库 `context.Context` 创建带应用名称的上下文
- 推荐使用 `Context.WithAppName()` 方法进行链式调用

#### WithTraceId
设置或生成追踪ID（包级别函数，用于从标准库 context 创建）：
```go
func WithTraceId(parent context.Context, traceId ...string) *Context
```
- 如果不提供 `traceId`，会自动生成 UUID
- 用于从标准库 `context.Context` 创建带追踪ID的上下文
- 推荐使用 `Context.WithTraceId()` 方法进行链式调用

#### AppName
获取应用名称：
```go
func (c *Context) AppName() string
```
- 支持多种 key 格式：`AppName`, `app-name`, `app_name`

#### TraceId
获取追踪ID：
```go
func (c *Context) TraceId() string
```
- 支持多种 key 格式：`TraceId`, `trace-id`, `trace_id`, `request_id`

#### FromGinContext
从 gin.Context 创建新上下文：
```go
func FromGinContext(ginCtx *gin.Context) *Context
```

#### FromGrpcContext
从 gRPC metadata 提取上下文信息：
```go
func FromGrpcContext(ctx context.Context) *Context
```

### 值获取方法（Context 方法）

#### ValueAny
获取原始值（需要手动类型断言）：
```go
func (c *Context) ValueAny(key string) any
```

#### ValueString
获取字符串类型的值，支持自动转换：
```go
func (c *Context) ValueString(key string) string
```
- 支持从 `int`, `int32`, `int64`, `float32`, `float64`, `bool` 自动转换

#### ValueInt
获取 int 类型的值，支持自动转换：
```go
func (c *Context) ValueInt(key string) int
```
- 支持从 `string`, `int32`, `int64`, `float32`, `float64`, `bool` 自动转换

#### ValueInt32
获取 int32 类型的值，支持自动转换：
```go
func (c *Context) ValueInt32(key string) int32
```
- 支持从 `string`, `int`, `int64`, `float32`, `float64`, `bool` 自动转换

#### ValueInt64
获取 int64 类型的值，支持自动转换：
```go
func (c *Context) ValueInt64(key string) int64
```
- 支持从 `string`, `int`, `int32`, `float32`, `float64`, `bool` 自动转换

#### ValueFloat32
获取 float32 类型的值，支持自动转换：
```go
func (c *Context) ValueFloat32(key string) float32
```
- 支持从 `string`, `int`, `int32`, `int64`, `float64`, `bool` 自动转换

#### ValueFloat64
获取 float64 类型的值，支持自动转换：
```go
func (c *Context) ValueFloat64(key string) float64
```
- 支持从 `string`, `int`, `int32`, `int64`, `float32`, `bool` 自动转换

#### ValueBool
获取 bool 类型的值，支持自动转换：
```go
func (c *Context) ValueBool(key string) bool
```
- 支持从 `string`, `int`, `int32`, `int64`, `float32`, `float64` 自动转换
- 对于 `string` 类型：
  - 返回 `false`：空字符串、`"0"`、`"nil"`、`"null"`、`"false"`、`"no"`、`"off"`（不区分大小写）
  - 其他情况返回 `true`
- 对于数值类型：非 0 返回 `true`，0 返回 `false`

## 类型转换规则

### ValueInt64 转换规则
- `string` → 使用 `strconv.ParseInt` 解析
- `int` → 直接转换
- `int32` → 直接转换
- `float32` → 截断小数部分
- `float64` → 截断小数部分
- `bool` → `true` 转为 1，`false` 转为 0

### ValueString 转换规则
- `int` → 使用 `strconv.FormatInt` 格式化
- `int32` → 使用 `strconv.FormatInt` 格式化
- `int64` → 使用 `strconv.FormatInt` 格式化
- `float32` → 使用 `strconv.FormatFloat` 格式化
- `float64` → 使用 `strconv.FormatFloat` 格式化
- `bool` → `true` 转为 `"true"`，`false` 转为 `"false"`

### ValueBool 转换规则
- `string` → 空字符串、`"0"`、`"nil"`、`"null"`、`"false"`、`"no"`、`"off"` 返回 `false`，其他返回 `true`（不区分大小写）
- `int` → 非 0 返回 `true`，0 返回 `false`
- `int32` → 非 0 返回 `true`，0 返回 `false`
- `int64` → 非 0 返回 `true`，0 返回 `false`
- `float32` → 非 0 返回 `true`，0 返回 `false`
- `float64` → 非 0 返回 `true`，0 返回 `false`

## 使用建议

### 推荐使用方式

1. **链式调用**: 使用 Context 方法进行链式调用，代码更简洁流畅：
   ```go
   ctx := goocontext.Default(context.Background()).
       WithAppName("my-app").
       WithTraceId().
       WithValue("user-id", "12345")
   ```

2. **获取值**: 使用 Context 方法获取值：
   ```go
   appName := ctx.AppName()
   traceId := ctx.TraceId()
   userId := ctx.ValueString("user-id")
   ```

3. **从标准库 context 创建**: 使用包级别函数：
   ```go
   ctx := goocontext.WithAppName(context.Background(), "my-app")
   ```

## 注意事项

1. **上下文不可变性**: 所有 `With*` 方法都会返回新的 `Context` 实例，不会修改原上下文
2. **类型转换失败**: 如果类型转换失败，会返回目标类型的零值（0、空字符串、false 等）
3. **nil 安全**: 所有方法都进行了 nil 检查，可以安全地传入 nil 上下文
4. **信号处理**: `WithSignalNotify` 会在后台启动 goroutine 监听信号，需要注意资源清理
5. **Gin 集成**: `WithGinContext` 会自动将 `app-name` 和 `trace-id` 设置到 gin.Context 中，方便在中间件中使用
6. **gRPC 集成**: `WithGrpcContext` 会将上下文信息添加到 gRPC metadata，需要在客户端和服务端都正确处理
7. **TraceId 格式**: 默认使用 UUID v4 格式生成 TraceId，也可以通过 `WithTraceId` 设置自定义格式
8. **方法调用**: 所有 `Value*` 方法都是 Context 的方法，使用 `ctx.ValueString("key")` 而不是包级别函数
