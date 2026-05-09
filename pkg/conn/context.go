package conn

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/Tudyha/nexus/pkg/proto"
)

type MessageHandler interface {
	Type() proto.MessageType
	Handle(ctx Context) error
}

type Context interface {
	context.Context
	GetConn() *Conn                      // 获取当前连接
	Unmarshal(v any) error               // 反序列化消息
	Hijack() (io.ReadWriteCloser, error) // 劫持连接，返回底层的 net.Conn，调用方负责关闭
	IsHijacked() bool                    // 是否被劫持
	WithValue(key any, value any)        // 设置上下文值
}

type connContext struct {
	context.Context
	conn     *Conn
	message  *proto.Message
	hijacked atomic.Bool
}

func NewConnContext(ctx context.Context, conn *Conn, message *proto.Message) Context {
	c := new(connContext)
	c.Context = ctx
	c.conn = conn
	c.message = message
	return c
}

func (c *connContext) Unmarshal(v any) error {
	return c.conn.Unmarshal(c.message.Payload, v)
}

func (c *connContext) Hijack() (io.ReadWriteCloser, error) {
	if c.hijacked.Load() {
		return nil, fmt.Errorf("connection already hijacked")
	}
	c.hijacked.Store(true)
	return c.conn, nil
}

func (c *connContext) IsHijacked() bool {
	return c.hijacked.Load()
}

func (c *connContext) GetConn() *Conn {
	return c.conn
}

func (c *connContext) WithValue(key any, value any) {
	c.Context = context.WithValue(c.Context, key, value)
}
