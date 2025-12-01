package goocos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Client COS 客户端封装
type Client struct {
	name   string
	config *Config
	client *cos.Client
}

// NewClient 创建新的 COS 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证必填字段
	if config.SecretID == "" {
		return nil, ErrEmptySecretID
	}
	if config.SecretKey == "" {
		return nil, ErrEmptySecretKey
	}
	if config.Region == "" && config.BaseURL == "" {
		return nil, ErrEmptyRegion
	}
	if config.Bucket == "" && config.BaseURL == "" {
		return nil, ErrEmptyBucket
	}

	// 转换为 COS SDK 的配置
	baseURL, opts := config.toCOSOptions()

	// 获取认证信息
	secretID, secretKey, sessionToken := config.getCredentials()

	// 创建客户端
	client := cos.NewClient(baseURL, &http.Client{
		Timeout: opts.Timeout,
		Transport: &cos.AuthorizationTransport{
			SecretID:     secretID,
			SecretKey:    secretKey,
			SessionToken: sessionToken,
		},
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Service.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to COS: %w", err)
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

// Client 获取 COS 客户端
func (c *Client) Client() *cos.Client {
	return c.client
}

// Close 关闭客户端连接（COS 客户端通常不需要显式关闭）
func (c *Client) Close() error {
	// COS SDK 的 HTTP 客户端会自动管理连接，无需显式关闭
	return nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Service.Get(ctx)
	return err
}

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error) {
	return c.client.Object.Put(ctx, key, r, opt)
}

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, key string, opt *cos.ObjectGetOptions) (*cos.Response, error) {
	return c.client.Object.Get(ctx, key, opt)
}

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, key string) (*cos.Response, error) {
	return c.client.Object.Delete(ctx, key)
}

// HeadObject 获取对象元信息
func (c *Client) HeadObject(ctx context.Context, key string, opt *cos.ObjectHeadOptions) (*cos.Response, error) {
	return c.client.Object.Head(ctx, key, opt)
}

// CopyObject 复制对象
func (c *Client) CopyObject(ctx context.Context, key, sourceURL string, opt *cos.ObjectCopyOptions) (*cos.CopyObjectResult, *cos.Response, error) {
	return c.client.Object.Copy(ctx, key, sourceURL, opt)
}

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, prefix string, opt *cos.BucketGetOptions) (*cos.BucketGetResult, *cos.Response, error) {
	if opt == nil {
		opt = &cos.BucketGetOptions{}
	}
	if prefix != "" {
		opt.Prefix = prefix
	}
	return c.client.Bucket.Get(ctx, opt)
}

// PutObjectACL 设置对象 ACL
func (c *Client) PutObjectACL(ctx context.Context, key string, opt *cos.ObjectPutACLOptions) (*cos.Response, error) {
	return c.client.Object.PutACL(ctx, key, opt)
}

// GetObjectACL 获取对象 ACL
func (c *Client) GetObjectACL(ctx context.Context, key string) (*cos.AccessControlPolicy, *cos.Response, error) {
	return c.client.Object.GetACL(ctx, key)
}

// InitiateMultipartUpload 初始化分片上传
func (c *Client) InitiateMultipartUpload(ctx context.Context, key string, opt *cos.InitiateMultipartUploadOptions) (*cos.InitiateMultipartUploadResult, *cos.Response, error) {
	return c.client.Object.InitiateMultipartUpload(ctx, key, opt)
}

// UploadPart 上传分片
func (c *Client) UploadPart(ctx context.Context, key string, uploadID string, partNumber int, r io.Reader, opt *cos.ObjectUploadPartOptions) (*cos.Response, error) {
	return c.client.Object.UploadPart(ctx, key, uploadID, partNumber, r, opt)
}

// CompleteMultipartUpload 完成分片上传
func (c *Client) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, opt *cos.CompleteMultipartUploadOptions) (*cos.CompleteMultipartUploadResult, *cos.Response, error) {
	return c.client.Object.CompleteMultipartUpload(ctx, key, uploadID, opt)
}

// AbortMultipartUpload 取消分片上传
func (c *Client) AbortMultipartUpload(ctx context.Context, key string, uploadID string) (*cos.Response, error) {
	return c.client.Object.AbortMultipartUpload(ctx, key, uploadID)
}

// ListMultipartUploads 列出进行中的分片上传
func (c *Client) ListMultipartUploads(ctx context.Context, opt *cos.BucketListMultipartUploadsOptions) (*cos.ListMultipartUploadsResult, *cos.Response, error) {
	return c.client.Bucket.ListMultipartUploads(ctx, opt)
}

