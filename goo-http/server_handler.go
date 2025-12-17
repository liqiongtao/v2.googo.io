package goohttp

import (
	"github.com/gin-gonic/gin"
)

type HandlerFunc func(*Context)

func wrapHandlers(handlers ...HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{Context: c}

		for _, handler := range handlers {
			handler(ctx)

			if c.IsAborted() {
				return
			}
		}
	}
}

func (s *Server) Get(path string, handlers ...HandlerFunc) {
	s.engine.GET(path, wrapHandlers(handlers...))
}

func (s *Server) Post(path string, handlers ...HandlerFunc) {
	s.engine.POST(path, wrapHandlers(handlers...))
}

func (s *Server) Put(path string, handlers ...HandlerFunc) {
	s.engine.PUT(path, wrapHandlers(handlers...))
}

func (s *Server) Delete(path string, handlers ...HandlerFunc) {
	s.engine.DELETE(path, wrapHandlers(handlers...))
}

func (s *Server) Patch(path string, handlers ...HandlerFunc) {
	s.engine.PATCH(path, wrapHandlers(handlers...))
}

func (s *Server) Options(path string, handlers ...HandlerFunc) {
	s.engine.OPTIONS(path, wrapHandlers(handlers...))
}

func (s *Server) Static(path, root string) {
	s.engine.Static(path, root)
}

func (s *Server) StaticFile(path, filepath string) {
	s.engine.StaticFile(path, filepath)
}

type RouterGroup struct {
	group *gin.RouterGroup
}

func (s *Server) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		group: s.engine.Group(path, wrapHandlers(handlers...)),
	}
}

func (rg *RouterGroup) Get(path string, handlers ...HandlerFunc) {
	rg.group.GET(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Post(path string, handlers ...HandlerFunc) {
	rg.group.POST(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Put(path string, handlers ...HandlerFunc) {
	rg.group.PUT(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Delete(path string, handlers ...HandlerFunc) {
	rg.group.DELETE(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Patch(path string, handlers ...HandlerFunc) {
	rg.group.PATCH(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Options(path string, handlers ...HandlerFunc) {
	rg.group.OPTIONS(path, wrapHandlers(handlers...))
}

func (rg *RouterGroup) Static(path, root string) {
	rg.group.Static(path, root)
}

func (rg *RouterGroup) StaticFile(path, filepath string) {
	rg.group.StaticFile(path, filepath)
}
