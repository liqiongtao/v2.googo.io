# goo-etcd

## 需求

1. 开发语言：golang
2. 包名: gooetcd
3. 目录: goo-etcd
4. 功能需求:

   * 定义Client对象，基于 go.etcd.io/etcd/client/v3
   * 定义Config对象
   * 要支持服务注册、解析
   * 要支持watch
   * 定义包方法

## 功能特性

- ✅ 基于 `go.etcd.io/etcd/client/v3` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 etcd 连接）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 支持连接超时、TLS 等配置
- ✅ 支持键值操作（Put、Get、Delete）
- ✅ 支持 Watch 监听
- ✅ 支持租约（Lease）操作
- ✅ 支持事务（Txn）操作
- ✅ 支持集群管理操作

## 快速开始

### 安装依赖

```bash
go get go.etcd.io/etcd/client/v3
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
  
    "v2.googo.io/goo-etcd"
)

func main() {
    // 1. 注册默认客户端
    config := gooetcd.DefaultConfig()
    config.Endpoints = []string{"localhost:2379"}
    config.Username = "root"
    config.Password = "password"
  
    if err := gooetcd.Register("default", config); err != nil {
        panic(err)
    }
  
    // 2. 使用默认客户端
    client, _ := gooetcd.Default()
    etcdClient := client.Client()
  
    ctx := context.Background()
  
    // 写入键值对
    _, err := etcdClient.Put(ctx, "key", "value")
    if err != nil {
        panic(err)
    }
  
    // 获取键值对
    resp, err := etcdClient.Get(ctx, "key")
    if err != nil {
        panic(err)
    }
  
    for _, ev := range resp.Kvs {
        fmt.Printf("key: %s, value: %s\n", ev.Key, ev.Value)
    }
  
    // 3. 使用包级别便捷方法
    defaultClient, _ := gooetcd.GetDefaultClient()
    defaultClient.Put(ctx, "key2", "value2")
}
```

### 多 Client 切换

```go
// 注册多个客户端
gooetcd.Register("cluster1", &gooetcd.Config{
    Endpoints: []string{"cluster1-etcd:2379"},
    Username:  "user1",
    Password:  "pass1",
})

gooetcd.Register("cluster2", &gooetcd.Config{
    Endpoints: []string{"cluster2-etcd:2379"},
    Username:  "user2",
    Password:  "pass2",
})

// 切换使用不同的客户端
cluster1Client, _ := gooetcd.GetClient("cluster1")
cluster2Client, _ := gooetcd.GetClient("cluster2")

// 设置默认客户端
gooetcd.SetDefault("cluster1")
defaultClient, _ := gooetcd.Default()
```

### Watch 监听

```go
client, _ := gooetcd.Default()
ctx := context.Background()

// 监听键值变化
watchChan := client.Watch(ctx, "key", clientv3.WithPrefix())

for watchResp := range watchChan {
    for _, ev := range watchResp.Events {
        fmt.Printf("Type: %s, Key: %s, Value: %s\n", 
            ev.Type, ev.Kv.Key, ev.Kv.Value)
    }
}
```

### 租约操作

```go
client, _ := gooetcd.Default()
ctx := context.Background()

// 创建租约（TTL 为 10 秒）
leaseResp, err := client.Grant(ctx, 10)
if err != nil {
    panic(err)
}

leaseID := leaseResp.ID

// 使用租约写入键值对
_, err = client.Put(ctx, "key", "value", clientv3.WithLease(leaseID))
if err != nil {
    panic(err)
}

// 保持租约存活
keepAliveChan, err := client.KeepAlive(ctx, leaseID)
if err != nil {
    panic(err)
}

// 处理 keepalive 响应
go func() {
    for ka := range keepAliveChan {
        fmt.Printf("Lease %d kept alive, TTL: %d\n", ka.ID, ka.TTL)
    }
}()
```

### 事务操作

```go
client, _ := gooetcd.Default()
ctx := context.Background()

// 事务：如果 key1 的值是 "value1"，则设置 key2 为 "value2"
txn := client.Txn(ctx)
txn.If(clientv3.Compare(clientv3.Value("key1"), "=", "value1")).
    Then(clientv3.OpPut("key2", "value2")).
    Else(clientv3.OpPut("key2", "value3"))

txnResp, err := txn.Commit()
if err != nil {
    panic(err)
}

if txnResp.Succeeded {
    fmt.Println("Transaction succeeded")
} else {
    fmt.Println("Transaction failed")
}
```

## API 文档

### Config 配置对象

```go
type Config struct {
    Endpoints            []string    // 端点列表，格式: ["localhost:2379"]
    Username             string      // 用户名
    Password             string      // 密码
    DialTimeout          time.Duration // 连接超时时间，默认 5 秒
    AutoSyncInterval     time.Duration // 自动同步间隔，默认 0
    AutoSync             bool        // 是否启用自动同步，默认 false
    MaxCallSendMsgSize   int         // 最大调用发送大小（字节），默认 2MB
    MaxCallRecvMsgSize   int         // 最大调用接收大小（字节），默认 2MB
    EnableCompression    bool        // 是否启用压缩，默认 false
    EnableGRPCDebugLog   bool        // 是否启用 gRPC 调试日志，默认 false
    EnableTLS            bool        // 是否启用客户端 TLS，默认 false
    TLSConfig            *TLSConfig  // TLS 配置
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 etcd 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取 etcd 客户端
func (c *Client) Client() *clientv3.Client

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error

// Put 写入键值对
func (c *Client) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)

// Get 获取键值对
func (c *Client) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)

// Delete 删除键值对
func (c *Client) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)

// Watch 监听键值变化
func (c *Client) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan

// Grant 创建租约
func (c *Client) Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error)

// Revoke 撤销租约
func (c *Client) Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)

// KeepAlive 保持租约存活
func (c *Client) KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error)

// KeepAliveOnce 保持租约存活一次
func (c *Client) KeepAliveOnce(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error)

// Txn 事务操作
func (c *Client) Txn(ctx context.Context) clientv3.Txn

// Status 获取集群状态
func (c *Client) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error)

// MemberList 获取成员列表
func (c *Client) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error)

// MemberAdd 添加成员
func (c *Client) MemberAdd(ctx context.Context, peerAddrs []string) (*clientv3.MemberAddResponse, error)

// MemberRemove 移除成员
func (c *Client) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error)

// MemberUpdate 更新成员
func (c *Client) MemberUpdate(ctx context.Context, id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error)
```

### 包级别方法

```go
// Register 注册一个 etcd 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个 etcd 客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultClient 获取默认客户端的 etcd 客户端
func GetDefaultClient() (*clientv3.Client, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 etcd 客户端
2. **连接管理**: 使用 `CloseAll()` 在应用退出时关闭所有连接
3. **多 Client 场景**: 当需要连接多个不同的 etcd 集群时，使用不同的名称注册多个客户端
4. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
5. **上下文使用**: 所有操作都使用 context，建议设置合理的超时时间
6. **Watch 使用**: Watch 操作会持续监听，注意合理管理 goroutine 和 context 取消
7. **租约使用**: 使用租约可以实现键值对的自动过期，适合实现分布式锁、服务注册等场景

## 注意事项

1. **依赖版本**: 本库基于 `go.etcd.io/etcd/client/v3`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **资源释放**: 使用完毕后记得调用 `Close()` 或 `CloseAll()` 释放连接资源
4. **默认客户端**: 如果没有设置默认客户端，使用 `Default()` 等方法会返回错误
5. **端点配置**: 建议配置多个端点以实现高可用，etcd 客户端会自动进行故障转移
6. **TLS 配置**: 生产环境建议启用 TLS 加密连接
7. **连接超时**: 根据网络环境合理设置 `DialTimeout`，避免连接时间过长
8. **消息大小**: 根据实际需求调整 `MaxCallSendMsgSize` 和 `MaxCallRecvMsgSize`，避免消息过大导致的问题
