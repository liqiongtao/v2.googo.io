package gooredis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Client Redis 客户端封装
type Client struct {
	name   string
	client *redis.Client
	dbs    map[int]*redis.Client // 多 db 支持
}

// NewClient 创建新的 Redis 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	opts := config.toRedisOptions()
	client := redis.NewClient(opts)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	c := &Client{
		name:   name,
		client: client,
		dbs:    make(map[int]*redis.Client),
	}

	// 将默认 db 添加到 dbs map
	c.dbs[config.DB] = client

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Client 获取默认数据库的客户端
func (c *Client) Client() *redis.Client {
	return c.client
}

// DB 获取指定数据库的客户端（支持多 db 选择）
func (c *Client) DB(db int) *redis.Client {
	// 如果已经存在，直接返回
	if client, ok := c.dbs[db]; ok {
		return client
	}

	// 创建新的客户端连接（使用相同的配置，但不同的 db）
	opts := c.client.Options()
	opts.DB = db
	newClient := redis.NewClient(opts)

	// 测试连接
	ctx := context.Background()
	if err := newClient.Ping(ctx).Err(); err != nil {
		// 如果连接失败，返回默认客户端
		return c.client
	}

	// 缓存到 dbs map
	c.dbs[db] = newClient
	return newClient
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	var err error
	for _, client := range c.dbs {
		if closeErr := client.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
