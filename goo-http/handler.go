package goohttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	goocontext "v2.googo.io/goo-context"
	goolog "v2.googo.io/goo-log"
)

// HandlerFunc 统一处理函数类型
type HandlerFunc func(*gin.Context) (interface{}, error)

// Handler 统一处理入口方法
func Handler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置traceId
		setupTraceId(c)

		// 处理请求体解密（如果启用）
		if err := decryptRequestBody(c); err != nil {
			Error(c, http.StatusBadRequest, "failed to decrypt request body: "+err.Error())
			return
		}

		// 记录请求日志
		logRequest(c)

		// 调用处理函数
		data, err := handler(c)

		// 处理响应
		if err != nil {
			handleError(c, err)
			return
		}

		// 成功响应
		Success(c, data)

		// 记录响应日志
		logResponse(c)

		// 处理响应体加密（如果启用）
		encryptResponseBody(c)
	}
}

// setupTraceId 设置traceId
func setupTraceId(c *gin.Context) {
	// 从请求头获取traceId
	traceId := c.GetHeader("X-Trace-Id")
	if traceId == "" {
		// 从gin.Context获取
		if v, exists := c.Get("trace-id"); exists {
			if id, ok := v.(string); ok {
				traceId = id
			}
		}
	}

	// 如果还没有traceId，生成一个
	if traceId == "" {
		ctx := goocontext.Default(c.Request.Context())
		ctx = ctx.WithTraceId()
		traceId = ctx.TraceId()
	}

	// 设置到gin.Context
	c.Set("trace-id", traceId)
	c.Header("X-Trace-Id", traceId)

	// 更新请求上下文
	ctx := goocontext.Default(c.Request.Context())
	ctx = ctx.WithTraceId(traceId)
	c.Request = c.Request.WithContext(ctx.Context)
}

// decryptRequestBody 解密请求体
func decryptRequestBody(c *gin.Context) error {
	// 检查是否启用加密
	server := getServerFromContext(c)
	if server == nil || server.config == nil || !server.config.Encryption.Enable {
		return nil
	}

	// 检查Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" && contentType != "text/plain" {
		return nil // 非JSON或文本类型，不处理
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	// 检查是否有加密标识
	// 这里假设加密的数据是Base64编码的字符串
	// 实际实现可能需要根据具体协议调整
	decryptor := server.config.Encryption.Decryptor
	if decryptor == nil {
		return nil
	}

	// 尝试解密
	decryptedBody, err := DecryptBase64(decryptor, string(body))
	if err != nil {
		// 如果解密失败，可能是未加密的数据，直接返回
		// 实际实现可能需要更严格的判断
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}

	// 替换请求体
	c.Request.Body = io.NopCloser(bytes.NewReader(decryptedBody))

	return nil
}

// encryptResponseBody 加密响应体
func encryptResponseBody(c *gin.Context) {
	// 检查是否启用加密
	server := getServerFromContext(c)
	if server == nil || server.config == nil || !server.config.Encryption.Enable {
		return
	}

	// 检查响应类型
	contentType := c.Writer.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" && contentType != "application/json" && contentType != "text/plain" {
		return // 非JSON或文本类型，不处理
	}

	// 获取响应体
	writer := c.Writer
	if w, ok := writer.(*responseWriter); ok {
		if w.body != nil && len(w.body) > 0 {
			encryptor := server.config.Encryption.Encryptor
			if encryptor != nil {
				encrypted, err := EncryptBase64(encryptor, w.body)
				if err == nil {
					// 更新响应体
					w.body = []byte(encrypted)
					writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(encrypted)))
				}
			}
		}
	}
}

// handleError 处理错误
func handleError(c *gin.Context, err error) {
	// 可以根据错误类型返回不同的状态码
	// 这里简化处理，统一返回500
	InternalServerError(c, err.Error())
}

// logRequest 记录请求日志
func logRequest(c *gin.Context) {
	server := getServerFromContext(c)
	if server == nil || server.config == nil || !server.config.Log.Enable {
		return
	}

	traceId := getTraceId(c)
	method := c.Request.Method
	path := c.Request.URL.Path
	ip := c.ClientIP()

	// 记录基本请求信息
	goolog.WithField("method", method).
		WithField("path", path).
		WithField("ip", ip).
		WithField("trace-id", traceId).
		InfoF("[goo-http] %s %s", method, path)

	// 记录请求体（如果启用）
	if server.config.Log.LogRequestBody {
		body, _ := io.ReadAll(c.Request.Body)
		if len(body) > 0 {
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			bodyStr := string(body)
			if server.config.Log.MaxRequestBodySize > 0 && len(bodyStr) > server.config.Log.MaxRequestBodySize {
				bodyStr = bodyStr[:server.config.Log.MaxRequestBodySize] + "..."
			}
			goolog.WithField("trace-id", traceId).
				DebugF("[goo-http] request body: %s", bodyStr)
		}
	}
}

// logResponse 记录响应日志
func logResponse(c *gin.Context) {
	server := getServerFromContext(c)
	if server == nil || server.config == nil || !server.config.Log.Enable {
		return
	}

	traceId := getTraceId(c)
	statusCode := c.Writer.Status()

	// 记录基本响应信息
	goolog.WithField("status", statusCode).
		WithField("trace-id", traceId).
		InfoF("[goo-http] response status: %d", statusCode)

	// 记录响应体（如果启用）
	if server.config.Log.LogResponseBody {
		if w, ok := c.Writer.(*responseWriter); ok {
			if w.body != nil && len(w.body) > 0 {
				bodyStr := string(w.body)
				if server.config.Log.MaxResponseBodySize > 0 && len(bodyStr) > server.config.Log.MaxResponseBodySize {
					bodyStr = bodyStr[:server.config.Log.MaxResponseBodySize] + "..."
				}
				goolog.WithField("trace-id", traceId).
					DebugF("[goo-http] response body: %s", bodyStr)
			}
		}
	}
}

// getServerFromContext 从gin.Context获取Server实例
func getServerFromContext(c *gin.Context) *Server {
	if v, exists := c.Get("goo-http-server"); exists {
		if server, ok := v.(*Server); ok {
			return server
		}
	}
	return nil
}

// responseWriter 包装gin.ResponseWriter以捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}

// wrapResponseWriter 包装响应写入器
func wrapResponseWriter(c *gin.Context) {
	if _, ok := c.Writer.(*responseWriter); !ok {
		c.Writer = &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
	}
}

// getTraceId 从gin.Context获取traceId
func getTraceId(c *gin.Context) string {
	if v, exists := c.Get("trace-id"); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// 请求开始时间
func getStartTime(c *gin.Context) time.Time {
	if v, exists := c.Get("start-time"); exists {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

