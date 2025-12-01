# goo-redis 缓存库

## 需求

1. 开发语言：golang
2. 包名: gooredis
3. 目录: goo-redis
4. 功能需求:
   * 定义Client对象，基于 github.com/go-redis/redis
   * 定义Config对象
   * 定义包方法
   * 支持多db选择
   * 支持多Client切换

## 功能特性

- ✅ 基于 `github.com/go-redis/redis/v8` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 Redis 连接）
- ✅ 支持多 DB 选择（同一个 Client 可以访问不同的数据库）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 自动连接池管理
- ✅ 支持连接超时、读写超时等配置

## 快速开始

### 安装依赖

```bash
go get github.com/go-redis/redis/v8
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    
    "v2.googo.io/goo-redis"
)

func main() {
    // 1. 注册默认客户端
    config := gooredis.DefaultConfig()
    config.Addr = "localhost:6379"
    config.Password = "your-password"
    config.DB = 0
    
    if err := gooredis.Register("default", config); err != nil {
        panic(err)
    }
    
    // 2. 使用默认客户端
    client, _ := gooredis.Default()
    redisClient := client.Client()
    
    ctx := context.Background()
    redisClient.Set(ctx, "key", "value", 0)
    val, _ := redisClient.Get(ctx, "key").Result()
    fmt.Println(val) // 输出: value
    
    // 3. 使用指定数据库
    db1Client, _ := gooredis.DB(1)
    db1Client.Set(ctx, "key", "value-db1", 0)
    
    // 4. 使用包级别便捷方法
    defaultClient, _ := gooredis.GetDefaultClient()
    defaultClient.Set(ctx, "key2", "value2", 0)
}
```

### 多 Client 切换

```go
// 注册多个客户端
gooredis.Register("cache", &gooredis.Config{
    Addr:     "cache-redis:6379",
    Password: "cache-password",
    DB:       0,
})

gooredis.Register("session", &gooredis.Config{
    Addr:     "session-redis:6379",
    Password: "session-password",
    DB:       0,
})

// 切换使用不同的客户端
cacheClient, _ := gooredis.GetClient("cache")
sessionClient, _ := gooredis.GetClient("session")

// 设置默认客户端
gooredis.SetDefault("cache")
defaultClient, _ := gooredis.Default()
```

### 多 DB 选择

```go
// 获取默认客户端的 DB 0
db0, _ := gooredis.DB(0)

// 获取默认客户端的 DB 1
db1, _ := gooredis.DB(1)

// 从指定客户端获取不同 DB
client, _ := gooredis.GetClient("cache")
db2 := client.DB(2)
```

## API 文档

### Config 配置对象

```go
type Config struct {
    Addr            string        // 连接地址，格式: host:port
    Username        string        // 用户名（Redis 6.0+）
    Password        string        // 密码
    DB              int           // 数据库编号，默认 0
    PoolSize        int           // 连接池最大连接数，默认 10
    MinIdleConns    int           // 连接池最小空闲连接数，默认 5
    ConnMaxIdleTime time.Duration // 连接最大空闲时间，默认 5 分钟
    ConnMaxLifetime time.Duration // 连接最大生存时间，默认 30 分钟
    DialTimeout     time.Duration // 连接超时时间，默认 5 秒
    ReadTimeout     time.Duration // 读取超时时间，默认 3 秒
    WriteTimeout    time.Duration // 写入超时时间，默认 3 秒
    MaxRetries      int           // 最大重试次数，默认 3
    MinRetryBackoff time.Duration // 重试间隔，默认 128 毫秒
    MaxRetryBackoff time.Duration // 最大重试间隔，默认 512 毫秒
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 Redis 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取默认数据库的客户端
func (c *Client) Client() *redis.Client

// DB 获取指定数据库的客户端（支持多 db 选择）
func (c *Client) DB(db int) *redis.Client

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error
```

### 包级别方法

```go
// Register 注册一个 Redis 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// Unregister 注销一个 Redis 客户端
func Unregister(name string) error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// DefaultDB 获取默认客户端的指定数据库（支持多 db 选择）
func DefaultDB(db int) (*redis.Client, error)

// GetDefaultClient 获取默认客户端的默认数据库
func GetDefaultClient() (*redis.Client, error)

// DB 获取默认客户端的指定数据库
func DB(db int) (*redis.Client, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 Redis 客户端
2. **连接管理**: 使用 `CloseAll()` 在应用退出时关闭所有连接
3. **多 Client 场景**: 当需要连接多个不同的 Redis 服务器时，使用不同的名称注册多个客户端
4. **多 DB 场景**: 当需要访问同一个 Redis 服务器的不同数据库时，使用 `DB()` 方法
5. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值

## 注意事项

1. **依赖版本**: 本库基于 `github.com/go-redis/redis/v8`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **连接池**: 每个 Client 和 DB 都会维护自己的连接池，注意合理配置连接池大小
4. **资源释放**: 使用完毕后记得调用 `Close()` 或 `CloseAll()` 释放连接资源
5. **默认客户端**: 如果没有设置默认客户端，使用 `Default()` 等方法会返回错误
