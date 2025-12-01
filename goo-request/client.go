package goorequest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Request HTTP请求客户端
type Request struct {
	name   string
	config *Config
	client *http.Client
}

// NewRequest 创建新的请求客户端
func NewRequest(name string, config *Config) (*Request, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	if config.TLS != nil {
		transport.TLSClientConfig = config.TLS
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	r := &Request{
		name:   name,
		config: config,
		client: client,
	}

	return r, nil
}

// Name 获取客户端名称
func (r *Request) Name() string {
	return r.name
}

// Client 获取HTTP客户端
func (r *Request) Client() *http.Client {
	return r.client
}

// buildURL 构建完整URL
func (r *Request) buildURL(path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	if r.config.BaseURL != "" {
		return strings.TrimSuffix(r.config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/"), nil
	}

	return path, nil
}

// buildHeaders 构建请求头
func (r *Request) buildHeaders(customHeaders map[string]string) http.Header {
	headers := make(http.Header)

	// 先添加默认请求头
	for k, v := range r.config.Headers {
		headers.Set(k, v)
	}

	// 再添加自定义请求头（会覆盖默认的）
	for k, v := range customHeaders {
		headers.Set(k, v)
	}

	return headers
}

// doRequest 执行HTTP请求（带重试机制）
func (r *Request) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	maxRetries := r.config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// 重试前等待
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(r.config.RetryInterval):
			}
		}

		// 创建带上下文的请求
		reqWithCtx := req.WithContext(ctx)

		resp, err := r.client.Do(reqWithCtx)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 如果是最后一次尝试，直接返回错误
		if i == maxRetries {
			break
		}
	}

	return nil, lastErr
}

// Get 发送GET请求
func (r *Request) Get(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error) {
	return r.request(ctx, http.MethodGet, path, headers, params, nil, nil)
}

// Post 发送POST请求
func (r *Request) Post(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	return r.request(ctx, http.MethodPost, path, headers, nil, body, nil)
}

// Put 发送PUT请求
func (r *Request) Put(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	return r.request(ctx, http.MethodPut, path, headers, nil, body, nil)
}

// Delete 发送DELETE请求
func (r *Request) Delete(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	return r.request(ctx, http.MethodDelete, path, headers, nil, body, nil)
}

// Head 发送HEAD请求
func (r *Request) Head(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error) {
	return r.request(ctx, http.MethodHead, path, headers, params, nil, nil)
}

// Patch 发送PATCH请求
func (r *Request) Patch(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	return r.request(ctx, http.MethodPatch, path, headers, nil, body, nil)
}

// request 通用请求方法
func (r *Request) request(ctx context.Context, method, path string, headers map[string]string, params map[string]string, body interface{}, files map[string]string) (*http.Response, error) {
	// 构建URL
	fullURL, err := r.buildURL(path)
	if err != nil {
		return nil, err
	}

	// 添加查询参数
	if params != nil && len(params) > 0 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	var reqBody io.Reader
	reqHeaders := r.buildHeaders(headers)

	// 处理文件上传
	if files != nil && len(files) > 0 {
		bodyBuf := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuf)

		// 添加文件
		for fieldName, filePath := range files {
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
			if err != nil {
				return nil, fmt.Errorf("failed to create form file: %w", err)
			}

			// 使用流式传输处理大文件
			if _, err := io.Copy(part, file); err != nil {
				return nil, fmt.Errorf("failed to copy file content: %w", err)
			}
		}

		// 添加其他body数据（如果body是map[string]string）
		if bodyMap, ok := body.(map[string]string); ok {
			for k, v := range bodyMap {
				if err := writer.WriteField(k, v); err != nil {
					return nil, fmt.Errorf("failed to write field: %w", err)
				}
			}
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		reqBody = bodyBuf
		reqHeaders.Set("Content-Type", writer.FormDataContentType())
	} else if body != nil {
		// 处理普通body
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
			if reqHeaders.Get("Content-Type") == "" {
				reqHeaders.Set("Content-Type", "text/plain; charset=utf-8")
			}
		case []byte:
			reqBody = bytes.NewReader(v)
			if reqHeaders.Get("Content-Type") == "" {
				reqHeaders.Set("Content-Type", "application/octet-stream")
			}
		case io.Reader:
			reqBody = v
		default:
			// JSON编码
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
			if reqHeaders.Get("Content-Type") == "" {
				reqHeaders.Set("Content-Type", "application/json; charset=utf-8")
			}
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header = reqHeaders

	// 执行请求
	return r.doRequest(ctx, req)
}

// UploadFile 上传文件（支持大文件）
func (r *Request) UploadFile(ctx context.Context, path string, headers map[string]string, fieldName, filePath string, extraFields map[string]string) (*http.Response, error) {
	files := map[string]string{
		fieldName: filePath,
	}
	return r.request(ctx, http.MethodPost, path, headers, nil, extraFields, files)
}

// DownloadFile 下载文件到指定路径（支持大文件）
func (r *Request) DownloadFile(ctx context.Context, path string, headers map[string]string, params map[string]string, savePath string) error {
	resp, err := r.Get(ctx, path, headers, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status code: %d", resp.StatusCode)
	}

	// 创建保存目录
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 使用流式传输处理大文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadFileToWriter 下载文件到Writer（支持大文件）
func (r *Request) DownloadFileToWriter(ctx context.Context, path string, headers map[string]string, params map[string]string, writer io.Writer) error {
	resp, err := r.Get(ctx, path, headers, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status code: %d", resp.StatusCode)
	}

	// 使用流式传输处理大文件
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to writer: %w", err)
	}

	return nil
}

// Close 关闭客户端（HTTP客户端通常不需要显式关闭，但保留接口一致性）
func (r *Request) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}

