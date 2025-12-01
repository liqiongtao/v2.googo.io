package gooetcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client etcd 客户端封装
type Client struct {
	name   string
	config *Config
	client *clientv3.Client
}

// NewClient 创建新的 etcd 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证端点
	if len(config.Endpoints) == 0 {
		return nil, ErrEmptyEndpoints
	}

	// 转换为 etcd client v3 的配置
	etcdConfig, err := config.toEtcdConfig()
	if err != nil {
		return nil, err
	}

	// 创建客户端
	client, err := clientv3.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Status(ctx, etcdConfig.Endpoints[0])
	if err != nil {
		client.Close()
		return nil, err
	}

	c := &Client{
		name:   name,
		config: config,
		client: client,
	}

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Client 获取 etcd 客户端
func (c *Client) Client() *clientv3.Client {
	return c.client
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	if len(c.config.Endpoints) == 0 {
		return ErrEmptyEndpoints
	}
	_, err := c.client.Status(ctx, c.config.Endpoints[0])
	return err
}

// Put 写入键值对
func (c *Client) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return c.client.Put(ctx, key, val, opts...)
}

// Get 获取键值对
func (c *Client) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return c.client.Get(ctx, key, opts...)
}

// Delete 删除键值对
func (c *Client) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return c.client.Delete(ctx, key, opts...)
}

// Watch 监听键值变化
func (c *Client) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return c.client.Watch(ctx, key, opts...)
}

// Grant 创建租约
func (c *Client) Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	return c.client.Grant(ctx, ttl)
}

// Revoke 撤销租约
func (c *Client) Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	return c.client.Revoke(ctx, id)
}

// KeepAlive 保持租约存活
func (c *Client) KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return c.client.KeepAlive(ctx, id)
}

// KeepAliveOnce 保持租约存活一次
func (c *Client) KeepAliveOnce(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	return c.client.KeepAliveOnce(ctx, id)
}

// Txn 事务操作
func (c *Client) Txn(ctx context.Context) clientv3.Txn {
	return c.client.Txn(ctx)
}

// Status 获取集群状态
func (c *Client) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return c.client.Status(ctx, endpoint)
}

// MemberList 获取成员列表
func (c *Client) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	return c.client.MemberList(ctx)
}

// MemberAdd 添加成员
func (c *Client) MemberAdd(ctx context.Context, peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	return c.client.MemberAdd(ctx, peerAddrs)
}

// MemberRemove 移除成员
func (c *Client) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	return c.client.MemberRemove(ctx, id)
}

// MemberUpdate 更新成员
func (c *Client) MemberUpdate(ctx context.Context, id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error) {
	return c.client.MemberUpdate(ctx, id, peerAddrs)
}
