package goohttp

import (
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	SuccessCode    = 0         // 成功状态码
	SuccessMessage = "success" // 成功消息
)

type ResponseHook func(ctx *Context, resp *Response)

type Response struct {
	Code    int    `json:"code,omitempty"`     // 业务状态码
	Message string `json:"message,omitempty"`  // 响应消息
	Data    any    `json:"data,omitempty"`     // 响应数据
	TraceId string `json:"trace_id,omitempty"` // 追踪ID
}

func Success(ctx *Context, data any) *Response {
	return &Response{
		Code:    SuccessCode,
		Message: SuccessMessage,
		Data:    data,
		TraceId: ctx.TraceId(),
	}
}

func SuccessWithMessage(ctx *Context, message string, data interface{}) *Response {
	return &Response{
		Code:    SuccessCode,
		Message: message,
		Data:    data,
		TraceId: ctx.TraceId(),
	}
}

func Error(ctx *Context, code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    nil,
		TraceId: ctx.TraceId(),
	}
}

func ErrorWithData(ctx *Context, code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
		TraceId: ctx.TraceId(),
	}
}

// 响应钩子写入器
type hookResponseWriter struct {
	gin.ResponseWriter
	Response *Response
	mu       sync.Mutex
}

func (w *hookResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := json.Unmarshal(data, w.Response); err != nil {
		w.Response = &Response{
			Code:    SuccessCode,
			Message: SuccessMessage,
			Data:    string(data),
		}
	}

	return w.ResponseWriter.Write(data)
}

func (w *hookResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func ResponseHookMiddleware(hooks []ResponseHook) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{Context: c}

		writer := &hookResponseWriter{
			ResponseWriter: c.Writer,
			Response:       &Response{},
		}

		c.Writer = writer

		c.Next()

		for _, hook := range hooks {
			hook(ctx, writer.Response)
		}
	}
}
