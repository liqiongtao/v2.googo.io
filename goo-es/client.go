package gooes

import (
	"context"
	"sync"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// Client Elasticsearch 客户端封装
type Client struct {
	name   string
	config *Config
	client *elasticsearch.Client
	mu     sync.RWMutex
}

// NewClient 创建新的 Elasticsearch 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证地址
	if len(config.Addresses) == 0 {
		return nil, ErrEmptyAddresses
	}

	// 转换为 go-elasticsearch 的配置
	esConfig := config.toElasticsearchConfig()

	// 创建客户端
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx := context.Background()
	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
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

// Client 获取 Elasticsearch 客户端
func (c *Client) Client() *elasticsearch.Client {
	return c.client
}

// Close 关闭客户端连接（Elasticsearch 客户端通常不需要显式关闭）
func (c *Client) Close() error {
	// go-elasticsearch 客户端没有 Close 方法
	// 如果需要，可以在这里添加清理逻辑
	return nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.client.Ping(c.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return err
	}

	return nil
}

// Info 获取集群信息
func (c *Client) Info(ctx context.Context) (*esapi.Response, error) {
	return c.client.Info(c.client.Info.WithContext(ctx))
}

// Search 执行搜索
func (c *Client) Search(opts ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	return c.client.Search(opts...)
}

// Index 索引文档
func (c *Client) Index(index string, body interface{}, opts ...func(*esapi.IndexRequest)) (*esapi.Response, error) {
	req := c.client.Index(index)
	if body != nil {
		req = req.WithBody(body)
	}
	for _, opt := range opts {
		opt(req)
	}
	return req.Do(context.Background(), c.client)
}

// Get 获取文档
func (c *Client) Get(index, documentID string, opts ...func(*esapi.GetRequest)) (*esapi.Response, error) {
	req := c.client.Get(index, documentID)
	for _, opt := range opts {
		opt(req)
	}
	return req.Do(context.Background(), c.client)
}

// Delete 删除文档
func (c *Client) Delete(index, documentID string, opts ...func(*esapi.DeleteRequest)) (*esapi.Response, error) {
	req := c.client.Delete(index, documentID)
	for _, opt := range opts {
		opt(req)
	}
	return req.Do(context.Background(), c.client)
}

// Update 更新文档
func (c *Client) Update(index, documentID string, body interface{}, opts ...func(*esapi.UpdateRequest)) (*esapi.Response, error) {
	req := c.client.Update(index, documentID)
	if body != nil {
		req = req.WithBody(body)
	}
	for _, opt := range opts {
		opt(req)
	}
	return req.Do(context.Background(), c.client)
}

// Bulk 批量操作
func (c *Client) Bulk(body interface{}, opts ...func(*esapi.BulkRequest)) (*esapi.Response, error) {
	req := c.client.Bulk()
	if body != nil {
		req = req.WithBody(body)
	}
	for _, opt := range opts {
		opt(req)
	}
	return req.Do(context.Background(), c.client)
}
