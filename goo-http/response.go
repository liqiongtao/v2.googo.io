package goohttp

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	SuccessCode    = 0         // 成功状态码
	SuccessMessage = "success" // 成功消息
)

type Response struct {
	Code    int    `json:"code,omitempty"`     // 业务状态码
	Message string `json:"message,omitempty"`  // 响应消息
	Data    any    `json:"data,omitempty"`     // 响应数据
	TraceId string `json:"trace_id,omitempty"` // 追踪ID
}

func (r *Response) IsSuccess() bool {
	return r.Code == SuccessCode
}

func (r *Response) Byte() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func (r *Response) String() string {
	return string(r.Byte())
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

type ResponseHook func(ctx *Context, resp *Response)

// 响应钩子写入器
type hookResponseWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
	mu     sync.Mutex
}

func (w *hookResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer.Write(data)
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
			buffer:         &bytes.Buffer{},
		}

		c.Writer = writer

		c.Next()

		var rsp *Response
		if err := json.Unmarshal(writer.buffer.Bytes(), &rsp); err != nil {
			rsp = Error(ctx, 5001, err.Error())
		}

		for _, hook := range hooks {
			hook(ctx, rsp)
		}
	}
}
