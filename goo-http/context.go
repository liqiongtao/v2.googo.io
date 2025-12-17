package goohttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
}

func (c *Context) TraceId() string {
	return c.Context.GetString("trace-id")
}

func (c *Context) SetTraceId(traceId string) {
	c.Context.Set("trace-id", traceId)
}

func (c *Context) ClientIP() string {
	if v := c.Context.GetHeader("X-Real-Ip"); v != "" {
		return v
	}
	return c.Context.ClientIP()
}

func (c *Context) Success(data any) {
	c.Context.JSON(SuccessCode, Success(c, data))
}

func (c *Context) SuccessWithMessage(message string, data interface{}) {
	c.Context.JSON(http.StatusOK, SuccessWithMessage(c, message, data))
}

func (c *Context) Error(code int, message string) {
	c.Context.JSON(http.StatusOK, Error(c, code, message))
}

func (c *Context) ErrorWithData(code int, message string, data interface{}) {
	c.Context.JSON(http.StatusOK, ErrorWithData(c, code, message, data))
}

func (c *Context) ErrorWithStatus(httpStatus int, code int, message string) {
	c.Context.JSON(httpStatus, Error(c, code, message))
}

func (c *Context) Abort(httpStatus int, code int, message string) {
	c.Context.JSON(httpStatus, Error(c, code, message))
	c.Context.Abort()
}
