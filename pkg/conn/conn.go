package conn

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/google/uuid"
)

type Conn struct {
	netConn net.Conn    // 底层连接
	codec   proto.Codec // 消息编解码器

	closeOnce sync.Once
}

func NewConn(netConn net.Conn) *Conn {
	return &Conn{
		netConn: netConn,
		codec:   proto.NewCodec(),
	}
}

func (c *Conn) ReadMessage() (*proto.Message, error) {
	return c.codec.ReadMessage(c.netConn)
}

func (c *Conn) ReadMessageAck() (*proto.Response, error) {
	message, err := c.codec.ReadMessage(c.netConn)
	if err != nil {
		return nil, err
	}
	var response proto.Response
	if err := c.codec.Unmarshal(message.Payload, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Conn) WriteMessage(messageType proto.MessageType, data any) error {
	b, err := c.codec.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	msg := &proto.Message{
		Id:        uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      messageType,
		Payload:   b,
	}
	return c.codec.WriteMessage(c.netConn, msg)
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.netConn.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.netConn.Write(b)
}

func (c *Conn) Unmarshal(b []byte, v any) error {
	return c.codec.Unmarshal(b, v)
}

func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.netConn.Close()
	})
	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return c.netConn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.netConn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.netConn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.netConn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.netConn.SetWriteDeadline(t)
}
