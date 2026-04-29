package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/Tudyha/nexus/client/config"
	"github.com/Tudyha/nexus/client/handler"
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xtaci/smux/v2"
)

type Client struct {
	cfg             *config.Config
	session         *smux.Session
	connected       atomic.Bool
	heartbeatStream net.Conn // 心跳流

	handlers   map[proto.MessageType]conn.MessageHandler // 消息处理器
	sysHandler *handler.SysHandler
}

// NewClient 创建一个新的客户端
func NewClient(cfg *config.Config) *Client {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	if cfg.ReconnectInterval == 0 {
		cfg.ReconnectInterval = 10
	}
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 30
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 10
	}

	sysHandler := handler.NewSysHandler()
	tunnelHandler := handler.NewTunnelHandler()
	a := &Client{
		cfg:        cfg,
		handlers:   make(map[proto.MessageType]conn.MessageHandler),
		sysHandler: sysHandler,
	}
	a.handlers[tunnelHandler.Type()] = tunnelHandler
	return a
}

// Run 运行客户端
func (c *Client) Run(ctx context.Context) {
	log.Info().Any("config", c.cfg).Msg("启动客户端")

	reconnectInterval := time.Duration(c.cfg.ReconnectInterval) * time.Second
	healthCheckInterval := time.Duration(c.cfg.HeartbeatInterval) * time.Second

	for {
		// 断线重连
		if !c.connected.Load() {
			if err := c.connect(); err != nil {
				log.Error().Err(err).Dur("reconnect_in", reconnectInterval).Msg("连接失败，稍后重试")
				select {
				case <-ctx.Done():
					return
				case <-time.After(reconnectInterval):
					continue
				}
			}
		}

		// 连接成功
		select {
		case <-ctx.Done():
			c.cleanup()
			return
		case <-time.After(healthCheckInterval):
			// 健康检查: session 是否关闭
			if c.session == nil || c.session.IsClosed() {
				log.Warn().Msg("session 已关闭，准备重连")
				c.connected.Store(false)
				c.cleanup()
				continue
			}
			// 发送心跳包
			c.heartbeat()
		}
	}
}

// cleanup 关闭 session 资源
func (c *Client) cleanup() {
	if c.session != nil {
		c.session.Close()
		c.session = nil
		c.heartbeatStream = nil
	}
}

// connect 尝试连接服务器
func (c *Client) connect() error {
	var err error
	connectTimeout := time.Duration(c.cfg.ConnectTimeout) * time.Second

	netConn, err := net.DialTimeout("tcp", c.cfg.ServerAddr, connectTimeout)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func(cause error) {
		if cause != nil {
			netConn.Close()
		}
	}(err)

	// 握手建立连接
	if err = c.handshake(netConn); err != nil {
		return err
	}

	// 握手成功
	session, err := smux.Client(netConn, nil)
	if err != nil {
		return fmt.Errorf("smux: %w", err)
	}
	c.session = session
	c.connected.Store(true)

	go c.acceptStream()
	return nil

}

// handshake 握手
func (c *Client) handshake(netConn net.Conn) error {
	conn := conn.NewConn(netConn)

	// 签名参数
	nonce := utils.RandHex(16)
	ts := time.Now().Unix()
	sig := utils.SignConnectPayload(uint64(c.cfg.AppId), c.cfg.AppSecret, ts, nonce)

	// 发送握手消息
	info, err := c.sysHandler.Info()
	if err != nil {
		return err
	}
	handshakeMsg := &proto.HandshakeReq{
		AppId:      c.cfg.AppId,
		Timestamp:  ts,
		Nonce:      nonce,
		Signature:  sig,
		Version:    c.cfg.Version,
		ClientInfo: info,
	}

	if err := conn.WriteMessage(proto.MessageType_HANDSHAKE, handshakeMsg); err != nil {
		return fmt.Errorf("write handshake: %w", err)
	}

	// 读取握手响应
	ack, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read handshake ack: %w", err)
	}
	if ack.GetType() != proto.MessageType_HANDSHAKE_ACK {
		return fmt.Errorf("unexpected handshake ack type: %v", ack.Type)
	}

	//TODO: 验证握手响应中的状态码

	return nil
}

// acceptStream 接收子流
func (c *Client) acceptStream() {
	for {
		stream, err := c.session.AcceptStream()
		if err != nil {
			log.Error().Err(err).Msg("AcceptStream 出错，session 可能已关闭")
			c.connected.Store(false)
			c.cleanup()
			return
		}
		go c.handleStream(stream)
	}
}

// handleStream 处理子流，根据消息类型调用相应的处理函数
func (c *Client) handleStream(netConn net.Conn) {
	co := conn.NewConn(netConn)

	for {
		message, err := co.ReadMessage()
		if err != nil {
			co.Close()
			return
		}

		ctx := conn.NewConnContext(context.Background(), co, message)
		h := c.handlers[message.GetType()]
		if h == nil {
			log.Info().Int("MessageType", int(message.GetType())).Msg("unknown message type")
			continue
		}

		// 处理消息
		if err := h.Handle(ctx); err != nil {
			log.Error().Err(err).Msg("handle message error")
			continue
		}

		if ctx.IsHijacked() {
			// 内网穿透关键：如果子流被劫持，则不再处理后续消息，由劫持方负责数据传输
			return
		}
	}
}

// heartbeat 发送心跳包
func (c *Client) heartbeat() {
	if !c.connected.Load() || c.session == nil || c.session.IsClosed() {
		// 未连接，不发送心跳包
		log.Info().Msg("未连接，不发送心跳包")
		return
	}
	if c.heartbeatStream == nil {
		heartbeatStream, err := c.session.OpenStream()
		if err != nil {
			log.Error().Err(err).Msg("OpenStream 出错，session 可能已关闭")
			return
		}
		c.heartbeatStream = heartbeatStream
	}

	heartbeatReq, err := c.sysHandler.SystemStats()
	if err != nil {
		log.Error().Err(err).Msg("获取系统状态数据出错")
		return
	}

	conn := conn.NewConn(c.heartbeatStream)

	conn.WriteMessage(proto.MessageType_HEARTBEAT, heartbeatReq)

	log.Info().Msg("发送心跳包成功")
}
