package session

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Tudyha/nexus/client/config"
	"github.com/Tudyha/nexus/client/version"
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	smuxcfg "github.com/Tudyha/nexus/pkg/smux"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xtaci/smux/v2"
)

// Session 管理到服务端的 smux 多路复用连接。
// 内部使用读写锁保护并发访问，所有方法都是 goroutine-safe 的。
type Session struct {
	// 连接配置
	appId     int64
	appSecret string

	log zerolog.Logger

	mu       sync.RWMutex
	smuxSess *smux.Session

	closeCh chan struct{}
	closed  bool // 受 mu 保护，标记 Close 是否已执行
}

// New 创建一个新的 Session。
func New(cfg *config.Config) *Session {
	return &Session{
		appId:     cfg.AppId,
		appSecret: cfg.AppSecret,
		log:       log.With().Str("component", "session").Logger(),
	}
}

// Connect 建立 TCP 连接、完成握手认证、创建 smux session。
// 返回后可通过 AcceptStream() 接受来自服务端的子流。
func (s *Session) Connect(ctx context.Context, netConn net.Conn, info *proto.ClientInfo) error {
	remoteAddr := netConn.RemoteAddr().String()
	s.log.Info().Str("remote", remoteAddr).Str("hostname", info.Hostname).Msg("开始握手")

	if err := s.handshake(netConn, info); err != nil {
		return err
	}

	s.log.Info().Msg("握手成功, 创建 smux session")
	sess, err := smux.Client(netConn, smuxcfg.DefaultConfig())
	if err != nil {
		return fmt.Errorf("smux client: %w", err)
	}

	s.mu.Lock()
	s.smuxSess = sess
	s.closeCh = make(chan struct{})
	s.closed = false
	s.mu.Unlock()

	s.log.Info().Str("remote", remoteAddr).Msg("smux session 建立完成")
	return nil
}

// handshake 与服务端完成握手认证。
func (s *Session) handshake(netConn net.Conn, info *proto.ClientInfo) error {
	co := conn.NewConn(netConn)

	nonce := utils.RandHex(16)
	ts := time.Now().Unix()
	sig := utils.SignConnectPayload(uint64(s.appId), s.appSecret, ts, nonce)

	handshakeMsg := &proto.HandshakeReq{
		AppId:       s.appId,
		Timestamp:   ts,
		Nonce:       nonce,
		Signature:   sig,
		Version:     version.Version,
		VersionName: version.VersionName,
		ClientInfo:  info,
	}

	s.log.Debug().Int64("app_id", s.appId).Int32("version", version.Version).Msg("发送握手消息")
	if err := co.WriteMessage(proto.MessageType_HANDSHAKE, handshakeMsg); err != nil {
		return fmt.Errorf("write handshake: %w", err)
	}

	ack, err := co.ReadMessage()
	if err != nil {
		return fmt.Errorf("read handshake ack: %w", err)
	}
	if ack.GetType() != proto.MessageType_HANDSHAKE_ACK {
		return fmt.Errorf("unexpected handshake ack type: %v", ack.Type)
	}

	s.log.Info().Msg("握手成功, 收到服务端确认")
	return nil
}

// Close 关闭底层连接和 smux session。
// 内部使用 closed 标志保证幂等性，可安全并发调用。
// 活跃 stream 观察到 Done() 关闭后应尽快退出。
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	ch := s.closeCh
	sess := s.smuxSess
	s.closeCh = nil
	s.smuxSess = nil
	s.mu.Unlock()

	if ch != nil {
		close(ch)
	}
	if sess != nil {
		s.log.Info().Msg("关闭 smux session")
		sess.Close()
	}
	return nil
}

// Done 返回一个 channel，当 session 开始关闭（Close 被调用）时被关闭。
// 活跃 stream 可通过此 channel 感知关闭信号，主动退出。
func (s *Session) Done() <-chan struct{} {
	s.mu.RLock()
	ch := s.closeCh
	s.mu.RUnlock()
	if ch != nil {
		return ch
	}
	// session 未连接或已完全关闭，返回一个已关闭的 channel
	closed := make(chan struct{})
	close(closed)
	return closed
}

// AcceptStream 阻塞等待并接受一个新的 smux 子流。
// 当 session 关闭时会返回错误。
func (s *Session) AcceptStream() (net.Conn, error) {
	s.mu.RLock()
	sess := s.smuxSess
	s.mu.RUnlock()
	if sess == nil {
		return nil, fmt.Errorf("session not connected")
	}
	stream, err := sess.AcceptStream()
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// OpenStream 创建一个新的 smux 子流。
func (s *Session) OpenStream() (net.Conn, error) {
	s.mu.RLock()
	sess := s.smuxSess
	s.mu.RUnlock()
	if sess == nil {
		return nil, fmt.Errorf("session not connected")
	}
	return sess.OpenStream()
}

// IsClosed 返回 session 是否已关闭或未连接。
func (s *Session) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.smuxSess == nil || s.smuxSess.IsClosed()
}
