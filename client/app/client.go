package app

import (
	"context"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"darvaza.org/x/net/reconnect"
	"github.com/Tudyha/nexus/client/config"
	"github.com/Tudyha/nexus/client/handler"
	"github.com/Tudyha/nexus/client/session"
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

type Client struct {
	cfg *config.Config

	reconnector *reconnect.Client

	sess       *session.Session
	registry   *handler.Registry
	sysHandler *handler.SysHandler

	preRestart func()

	closeOnce sync.Once
}

type Options struct {
	Middleware []handler.Middleware
}

func NewClient(ctx context.Context, cfg *config.Config, opts Options) (*Client, error) {
	c := &Client{
		cfg:        cfg,
		sess:       session.New(cfg),
		registry:   handler.NewRegistry(),
		sysHandler: handler.NewSysHandler(),
	}

	// 自动重连处理器
	cli, err := reconnect.New(&reconnect.Config{
		Context:        ctx,
		Remote:         cfg.ServerAddr,
		DialTimeout:    time.Duration(cfg.ConnectTimeout) * time.Second,
		ReconnectDelay: time.Duration(cfg.ReconnectInterval) * time.Second,
		OnConnect:      c.onConnect,
		OnSession:      c.onSession,
		OnDisconnect:   c.onDisconnect,
		OnError:        c.onError,
	})
	c.reconnector = cli
	if err != nil {
		return nil, err
	}

	// 注册消息处理中间件
	for _, mw := range opts.Middleware {
		c.registry.Use(mw)
	}

	log.Info().Int("handler_count", c.registry.Len()).Msg("客户端初始化完成")
	return c, nil
}

func (c *Client) Run() error {
	log.Info().Str("server", c.cfg.ServerAddr).Msg("正在连接服务端...")
	if err := c.reconnector.Connect(); err != nil {
		return err
	}
	return c.reconnector.Wait()
}

func (c *Client) Shutdown() {
	log.Info().Msg("正在关闭客户端...")
	c.closeOnce.Do(func() {
		if c.sess != nil {
			c.sess.Close()
		}
		if c.reconnector != nil {
			c.reconnector.Shutdown(context.Background())
		}
	})
}

func (c *Client) onConnect(ctx context.Context, conn net.Conn) error {
	log.Info().Msg("连接成功, 开始创建session")
	info, err := c.sysHandler.Info()
	if err != nil {
		return err
	}
	return c.sess.Connect(ctx, conn, info)
}

func (c *Client) onSession(ctx context.Context) error {
	log.Info().Msg("session 建立成功, 开始接收 stream")

	go c.heartbeatLoop(ctx)

	for {
		stream, err := c.sess.AcceptStream()
		if err != nil {
			log.Warn().Err(err).Msg("AcceptStream 退出, session 结束")
			return err
		}
		log.Debug().Str("remote", stream.RemoteAddr().String()).Msg("收到新 stream")
		go c.handleStream(ctx, stream)
	}
}

func (c *Client) onDisconnect(ctx context.Context, conn net.Conn) error {
	log.Info().Msg("连接已断开, 关闭session")
	return c.sess.Close()
}

func (c *Client) onError(ctx context.Context, conn net.Conn, err error) error {
	log.Info().Err(err).Msg("连接异常")
	return err
}

func (c *Client) handleStream(ctx context.Context, s net.Conn) {
	co := conn.NewConn(s)

	for {
		message, err := co.ReadMessage()
		if err != nil {
			log.Debug().Err(err).Msg("stream 读取结束")
			co.Close()
			return
		}
		msgType := message.GetType()
		log.Debug().Int("type", int(msgType)).Str("type_name", msgType.String()).Msg("收到消息")
		connCtx := conn.NewConnContext(ctx, co, message)
		if err := c.registry.Handle(connCtx, message); err != nil {
			log.Warn().Err(err).Int("type", int(msgType)).Msg("handle message failed")
		}
		if connCtx.IsHijacked() {
			log.Debug().Msg("stream 被劫持, 不再处理后续消息")
			return
		}
	}
}

func (c *Client) heartbeatLoop(ctx context.Context) {
	interval := time.Duration(c.cfg.HeartbeatInterval) * time.Second
	log.Info().Dur("interval", interval).Msg("心跳 goroutine 启动")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer log.Info().Msg("心跳 goroutine 退出")

	for {
		select {
		case <-ticker.C:
			c.sendHeartbeat()
		case <-c.sess.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) sendHeartbeat() {
	stream, err := c.sess.OpenStream()
	if err != nil {
		log.Error().Err(err).Msg("open heartbeat stream failed")
		return
	}
	defer stream.Close()

	stats, err := c.sysHandler.SystemStats()
	if err != nil {
		log.Error().Err(err).Msg("collect heartbeat data failed")
		return
	}

	co := conn.NewConn(stream)
	defer co.Close()
	if err := co.WriteMessage(proto.MessageType_HEARTBEAT, stats); err != nil {
		log.Error().Err(err).Msg("send heartbeat failed")
		return
	}
	log.Debug().Msg("heartbeat sent")
}

func (c *Client) spawnNewProcess() {
	if c.preRestart != nil {
		c.preRestart()
	}
	exe, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("get executable path failed")
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("spawn new process failed")
		return
	}
	log.Info().Int("pid", cmd.Process.Pid).Msg("new process started")
}
