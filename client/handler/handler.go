package handler

import (
	"fmt"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
)

// Middleware 包装 conn.MessageHandler，用于添加横切关注点（日志、鉴权、指标等）。
// 返回的 handler 必须保持 Type() 不变。
type Middleware func(next conn.MessageHandler) conn.MessageHandler

// Registry 管理消息处理器注册和 stream 消息分发。
type Registry struct {
	handlers   map[proto.MessageType]conn.MessageHandler
	middleware []Middleware
}

func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[proto.MessageType]conn.MessageHandler),
	}
	// 隧道消息处理器
	r.Register(NewTunnelHandler())

	// 进程消息处理器
	r.Register(NewProcessListHandler())
	r.Register(NewProcessKillHandler())

	// 文件消息处理器
	r.Register(NewFileListHandler())
	r.Register(NewFileDownloadHandler())
	r.Register(NewFileUploadHandler())
	r.Register(NewFileDeleteHandler())
	r.Register(NewFileMkdirHandler())
	r.Register(NewFileRenameHandler())

	// 网络消息处理器
	r.Register(NewNetworkListHandler())
	return r
}

// Register 注册一个消息处理器。相同 MessageType 的后续注册会覆盖前者。
func (r *Registry) Register(h conn.MessageHandler) *Registry {
	r.handlers[h.Type()] = h
	return r
}

// Use 注册一个中间件。中间件按注册顺序从外到内包装 handler。
func (r *Registry) Use(mw Middleware) {
	r.middleware = append(r.middleware, mw)
}

// Len 返回已注册的处理器数量。
func (r *Registry) Len() int {
	return len(r.handlers)
}

// Handle 处理消息
func (r *Registry) Handle(ctx conn.Context, message *proto.Message) error {
	h := r.handlers[message.GetType()]
	if h == nil {
		return fmt.Errorf("unkown message type: %d", message.GetType())
	}

	// 应用中间件链
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}

	return h.Handle(ctx)
}
