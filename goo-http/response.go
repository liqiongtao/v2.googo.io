package goohttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
	TraceId string      `json:"trace_id,omitempty"` // 追踪ID
}

// ResponseHook 响应钩子函数类型
type ResponseHook func(c *gin.Context, resp *Response)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	
	// 添加traceId
	if traceId := getTraceId(c); traceId != "" {
		resp.TraceId = traceId
	}
	
	// 执行响应钩子
	executeResponseHooks(c, resp)
	
	c.JSON(http.StatusOK, resp)
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	resp := &Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
	
	// 添加traceId
	if traceId := getTraceId(c); traceId != "" {
		resp.TraceId = traceId
	}
	
	// 执行响应钩子
	executeResponseHooks(c, resp)
	
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	resp := &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	
	// 添加traceId
	if traceId := getTraceId(c); traceId != "" {
		resp.TraceId = traceId
	}
	
	// 执行响应钩子
	executeResponseHooks(c, resp)
	
	c.JSON(http.StatusOK, resp)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	resp := &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	
	// 添加traceId
	if traceId := getTraceId(c); traceId != "" {
		resp.TraceId = traceId
	}
	
	// 执行响应钩子
	executeResponseHooks(c, resp)
	
	c.JSON(http.StatusOK, resp)
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "bad request"
	}
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "forbidden"
	}
	Error(c, http.StatusForbidden, message)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "not found"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	Error(c, http.StatusInternalServerError, message)
}

// getTraceId 从gin.Context获取traceId
func getTraceId(c *gin.Context) string {
	if c == nil {
		return ""
	}
	
	// 先从gin.Context中获取
	if v, exists := c.Get("trace-id"); exists {
		if id, ok := v.(string); ok && id != "" {
			return id
		}
	}
	
	// 从请求头中获取
	if traceId := c.GetHeader("X-Trace-Id"); traceId != "" {
		return traceId
	}
	
	return ""
}

// executeResponseHooks 执行响应钩子
func executeResponseHooks(c *gin.Context, resp *Response) {
	// 从gin.Context中获取响应钩子列表
	if v, exists := c.Get("response-hooks"); exists {
		if hooks, ok := v.([]ResponseHook); ok {
			for _, hook := range hooks {
				if hook != nil {
					hook(c, resp)
				}
			}
		}
	}
}

