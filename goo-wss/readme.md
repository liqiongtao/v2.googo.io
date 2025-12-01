# goo-wss

基于 `github.com/gorilla/websocket` 的 WebSocket 服务封装库，提供客户端管理、房间管理和消息广播功能。

## 需求

1. 开发语言：golang
2. 包名: goowss
3. 目录: goo-wss
4. 功能需求:

   * 定义Client对象，基于 github.com/gorilla/websocket
   * 定义Config对象
   * 定义Room对象
   * 每个Room内可以发广播消息
   * 可以给指定Client发送消息
   * 定义包方法

## 功能特性

- ✅ **Client 管理**：基于 gorilla/websocket 的客户端封装，支持自动心跳、消息队列
- ✅ **Room 管理**：房间机制，支持客户端分组管理
- ✅ **消息广播**：支持向房间内所有客户端广播消息
- ✅ **单播消息**：支持向指定客户端发送消息
- ✅ **Hub 管理器**：统一的客户端和房间管理中心
- ✅ **配置灵活**：丰富的配置选项，支持自定义超时、缓冲区大小等
- ✅ **线程安全**：所有操作都经过互斥锁保护，支持并发访问

## 快速开始

### 安装

```bash
go get github.com/your-org/goo-wss
```

### 基本使用

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/gorilla/websocket"
    "github.com/your-org/goo-wss"
)

func main() {
    // 创建配置
    config := goowss.DefaultConfig(
        goowss.WithReadBufferSize(4096),
        goowss.WithWriteBufferSize(4096),
        goowss.WithPongWait(60*time.Second),
        goowss.WithPingPeriod(54*time.Second),
    )
    
    // 创建 Hub
    hub := goowss.NewHub(config)
    
    // HTTP 处理器
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        // 升级连接
        conn, err := hub.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("升级连接失败: %v", err)
            return
        }
        
        // 创建客户端（使用唯一 ID）
        clientID := generateClientID() // 你需要实现这个函数
        client := goowss.NewClient(clientID, conn, config)
        
        // 注册客户端
        hub.RegisterClient(client)
        
        // 加入房间
        room := hub.GetOrCreateRoom("room1")
        room.AddClient(client)
        
        // 启动读写循环
        go client.WritePump()
        go client.ReadPump(
            func(c *goowss.Client, msg []byte) {
                // 处理接收到的消息
                log.Printf("收到消息: %s", string(msg))
                
                // 广播到房间
                hub.BroadcastToRoom("room1", msg)
            },
            func(c *goowss.Client, err error) {
                // 处理错误
                log.Printf("客户端错误: %v", err)
                hub.UnregisterClient(c.GetID())
            },
        )
    })
    
    log.Println("WebSocket 服务器启动在 :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## API 文档

### Config

WebSocket 配置对象。

```go
type Config struct {
    ReadBufferSize     int
    WriteBufferSize    int
    EnableCompression  bool
    CheckOrigin        func(r *http.Request) bool
    HandshakeTimeout   time.Duration
    ReadTimeout        time.Duration
    WriteTimeout       time.Duration
    PongWait           time.Duration
    PingPeriod         time.Duration
    MaxMessageSize     int64
    Upgrader           *websocket.Upgrader
}
```

**方法：**
- `DefaultConfig(opts ...FuncOption) *Config` - 创建默认配置

### Client

WebSocket 客户端对象。

**方法：**
- `NewClient(id string, conn *websocket.Conn, config *Config) *Client` - 创建新客户端
- `GetID() string` - 获取客户端 ID
- `GetRoomID() string` - 获取所属房间 ID
- `Send(message []byte) error` - 发送消息
- `ReadPump(onMessage func(*Client, []byte), onError func(*Client, error))` - 读取消息循环
- `WritePump()` - 写入消息循环
- `Close() error` - 关闭连接
- `IsClosed() bool` - 检查是否已关闭

### Room

WebSocket 房间对象。

**方法：**
- `NewRoom(id string) *Room` - 创建新房间
- `AddClient(client *Client) error` - 添加客户端
- `RemoveClient(clientID string) error` - 移除客户端
- `GetClient(clientID string) (*Client, error)` - 获取客户端
- `Broadcast(message []byte)` - 广播消息
- `BroadcastExclude(message []byte, excludeClientID string)` - 广播消息（排除指定客户端）
- `SendToClient(clientID string, message []byte) error` - 向指定客户端发送消息
- `GetClientCount() int` - 获取客户端数量
- `GetClients() []string` - 获取所有客户端 ID
- `Close()` - 关闭房间

### Hub

WebSocket 中心管理器。

**方法：**
- `NewHub(config *Config) *Hub` - 创建新 Hub
- `Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)` - 升级 HTTP 连接
- `RegisterClient(client *Client)` - 注册客户端
- `UnregisterClient(clientID string)` - 注销客户端
- `GetOrCreateRoom(roomID string) *Room` - 获取或创建房间
- `GetRoom(roomID string) (*Room, error)` - 获取房间
- `RemoveRoom(roomID string)` - 移除房间
- `GetClient(clientID string) (*Client, error)` - 获取客户端
- `SendToClient(clientID string, message []byte) error` - 向指定客户端发送消息
- `BroadcastToRoom(roomID string, message []byte) error` - 向房间广播消息
- `BroadcastToRoomExclude(roomID string, message []byte, excludeClientID string) error` - 向房间广播消息（排除指定客户端）
- `GetRoomCount() int` - 获取房间数量
- `GetClientCount() int` - 获取客户端数量

### 配置选项

- `WithReadBufferSize(size int) FuncOption` - 设置读取缓冲区大小
- `WithWriteBufferSize(size int) FuncOption` - 设置写入缓冲区大小
- `WithEnableCompression(enable bool) FuncOption` - 设置是否启用压缩
- `WithCheckOrigin(checkOrigin func(r *http.Request) bool) FuncOption` - 设置来源检查函数
- `WithHandshakeTimeout(timeout time.Duration) FuncOption` - 设置握手超时时间
- `WithReadTimeout(timeout time.Duration) FuncOption` - 设置读取超时时间
- `WithWriteTimeout(timeout time.Duration) FuncOption` - 设置写入超时时间
- `WithPongWait(duration time.Duration) FuncOption` - 设置 Pong 等待时间
- `WithPingPeriod(duration time.Duration) FuncOption` - 设置 Ping 周期
- `WithMaxMessageSize(size int64) FuncOption` - 设置最大消息大小
- `WithUpgrader(upgrader *websocket.Upgrader) FuncOption` - 设置自定义升级器

## 使用建议

1. **客户端 ID 生成**：建议使用 UUID 或其他唯一标识符作为客户端 ID
2. **错误处理**：务必处理 `ReadPump` 和 `WritePump` 中的错误，及时清理资源
3. **房间管理**：根据业务需求合理使用房间机制，避免创建过多房间
4. **消息大小**：根据实际需求设置合适的 `MaxMessageSize`，避免内存溢出
5. **超时设置**：合理设置 `PongWait` 和 `PingPeriod`，建议 `PingPeriod` 为 `PongWait` 的 90%
6. **并发安全**：所有操作都是线程安全的，可以在多个 goroutine 中安全使用

## 注意事项

1. **资源清理**：客户端断开连接后，务必调用 `UnregisterClient` 清理资源
2. **房间清理**：不再使用的房间应该调用 `RemoveRoom` 进行清理
3. **消息格式**：默认使用文本消息格式，如需二进制消息，需要修改 `WritePump` 方法
4. **跨域问题**：生产环境需要设置 `CheckOrigin` 函数，防止跨域攻击
5. **心跳机制**：客户端会自动发送 Ping 消息，服务端会自动响应 Pong，无需手动处理
6. **消息队列**：`Send` 方法使用缓冲通道，如果通道已满会返回错误，需要处理这种情况
7. **连接关闭**：客户端关闭时，`ReadPump` 和 `WritePump` 会自动退出，但需要确保在 `onError` 回调中清理资源
