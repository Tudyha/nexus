package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tudyha/nexus/pkg/conn"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	smux "github.com/xtaci/smux/v2"
)

const (
	StatusInit    int32 = 0
	StatusPending int32 = 1 // Awaiting authentication
	StatusReady   int32 = 2 // Authenticated and smux ready
	StatusClosed  int32 = 3
)

// 客户端会话
type Session struct {
	Id      string        //会话id
	session *smux.Session // smux session
	status  atomic.Int32  // session状态: 0: 初始化 1: 待认证 2: 认证成功 3: 关闭

	mu      sync.Mutex
	closed  bool
	closeCh chan struct{}
}

// newSession 创建一个新的 Session，初始状态为 StatusPending，并启动认证检查协程
func newSession(netConn net.Conn) *Session {
	s := &Session{
		Id:      uuid.New().String(),
		session: nil,
		closed:  false,
		closeCh: make(chan struct{}),
	}

	// 初始状态为待认证
	s.status.Store(StatusPending)

	// 启动认证检查协程
	go s.checkAuth(netConn)
	return s
}

// checkAuth 验证客户端身份，成功后将状态改为 StatusReady，并初始化 smux session
func (s *Session) checkAuth(netConn net.Conn) {
	var err error

	c := conn.NewConn(netConn)
	defer func() {
		if err != nil || s.status.Load() != StatusReady {
			log.Error().Err(err).Msg("authentication failed")
			c.Close()
			s.Close()
		}
	}()

	// 读握手消息，验证身份
	message, err := c.ReadMessage()
	if err != nil {
		return
	}
	// 判断消息类型，必须是握手消息
	if message.GetType() != proto.MessageType_HANDSHAKE {
		return
	}

	// 处理认证消息
	if err = s.handleMessage(c, message); err != nil {
		return
	}

	// 认证成功，发送握手响应
	if err = c.WriteMessage(proto.MessageType_HANDSHAKE_ACK, &proto.Response{}); err != nil {
		return
	}

	// 初始化 smux session（低延迟优化配置）
	s.session, err = smux.Server(netConn, &smux.Config{
		KeepAliveInterval: 5 * time.Second,  // 心跳间隔
		KeepAliveTimeout:  15 * time.Second, // 超时，需 ≥ Interval
		MaxFrameSize:      65535,            // 最大允许值（64KB）
		MaxReceiveBuffer:  2 * 1024 * 1024,  // 2MB
		MaxStreamBuffer:   1 * 1024 * 1024,  // 1MB，需 ≤ MaxReceiveBuffer
	})
	if err != nil {
		return
	}
	s.status.Store(StatusReady)

	// 监听新的 smux 流
	go s.acceptStream()
	log.Info().Str("SessionID", s.Id).Msg("authentication successful, session ready")
}

// handleMessage 处理来自客户端的消息
func (s *Session) handleMessage(c *conn.Conn, message *proto.Message) error {
	// 创建连接上下文
	connCtx := conn.NewConnContext(context.Background(), c, message)

	//通过context传递通用数据
	connCtx.WithValue(constant.ContextKeySessionId, s.Id)

	// 调用对应的消息处理器进行处理
	h, ok := messageHandlers[message.GetType()]
	if !ok {
		log.Error().Int("MessageType", int(message.GetType())).Msg("unknown message type")
		return fmt.Errorf("unknown message type: %v", message.GetType())
	}
	return h.Handle(connCtx)
}

// 接收 smux session 中的流，并为每个流启动一个新的协程处理
func (s *Session) acceptStream() {
	defer s.Close()
	for {
		conn, err := s.session.AcceptStream()
		if err != nil {
			// log.Error().Err(err).Msg("failed to accept stream")
			break
		}
		go s.handleStream(conn)
	}
}

// handleStream 处理 smux session 中的每个流
func (s *Session) handleStream(netConn net.Conn) {
	c := conn.NewConn(netConn)
	for {
		message, err := c.ReadMessage()
		if err != nil {
			c.Close()
			return
		}

		if err := s.handleMessage(c, message); err != nil {
			log.Error().Err(err).Msg("failed to handle message")
			continue
		}
	}
}

// Close 关闭 Session，释放资源
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		s.status.Store(StatusClosed)
		close(s.closeCh)
		if s.session != nil {
			s.session.Close()
		}
		// 删除 session
		managerInstance.sessions.Delete(s.Id)

		s.handleMessage(nil, &proto.Message{
			Type: proto.MessageType_DISCONNECT,
		})
	}
}

// openTunnel 打开一个隧道，返回一个 net.Conn 对象
func (s *Session) OpenTunnel(tunnelType proto.TunnelType, remoteAddr string) (net.Conn, error) {
	var err error
	if s.status.Load() != StatusReady {
		return nil, errcode.ErrClientNotReady
	}

	// 打开一个 smux 流
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, err
	}

	c := conn.NewConn(stream)
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	// 发送打开隧道请求
	req := &proto.TunnelOpenReq{
		Type:       tunnelType,
		RemoteAddr: remoteAddr,
	}

	if err = c.WriteMessage(proto.MessageType_TUNNEL_OPEN, req); err != nil {
		return nil, err
	}

	// 读取打开隧道响应
	ack, err := c.ReadMessageAck()
	if err != nil {
		return nil, err
	}

	// 隧道打开失败
	if ack.GetCode() != proto.ErrorCode_SUCCESS {
		return nil, errors.New(ack.GetMsg())
	}

	return stream, nil
}

func (s *Session) execute(msgType proto.MessageType) error {
	if s.status.Load() != StatusReady {
		return errcode.ErrClientNotReady
	}

	stream, err := s.session.OpenStream()
	if err != nil {
		return err
	}

	c := conn.NewConn(stream)
	defer c.Close()

	// 发请求
	if err := c.WriteMessage(msgType, &proto.ProcessReq{}); err != nil {
		return err
	}

	// 读响应
	_, err = c.ReadMessageAck()
	if err != nil {
		return err
	}

	// 反序列化响应
	return nil
}

func (s *Session) Exit() error {
	return s.execute(proto.MessageType_EXIT)
}

// SendTask 发送任务到客户端，在同一个 smux 流上读取进度上报
func (s *Session) SendTask(taskID uint64, taskType proto.TaskType, payload []byte, reportProgress bool, onProgress func(*proto.TaskProgress)) error {
	if s.status.Load() != StatusReady {
		return errcode.ErrClientNotReady
	}

	stream, err := s.session.OpenStream()
	if err != nil {
		return err
	}

	c := conn.NewConn(stream)
	defer c.Close()

	req := &proto.Task{
		TaskId:         taskID,
		TaskType:       taskType,
		ReportProgress: reportProgress,
		Payload:        payload,
	}

	if err := c.WriteMessage(proto.MessageType_TASK, req); err != nil {
		return err
	}

	// 循环读取 TASK_PROGRESS 直到 done=true 或 EOF
	for {
		msg, err := c.ReadMessage()
		if err != nil {
			// EOF 或连接断开，客户端已完成通信
			return nil
		}
		var progress proto.TaskProgress
		if err := c.Unmarshal(msg.Payload, &progress); err != nil {
			return err
		}
		if onProgress != nil {
			onProgress(&progress)
		}
		if progress.Done {
			return nil
		}
	}
}
