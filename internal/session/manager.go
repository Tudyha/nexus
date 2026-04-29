package session

import (
	"net"
	"sync"

	"github.com/Tudyha/nexus/internal/handler"
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

// session管理器
type Manager interface {
	NewSession(conn net.Conn) error
	OpenTunnel(sessionId string, tunnelType proto.TunnelType, targetAddr string) (net.Conn, error)
}

var (
	// 消息处理器，session不做消息处理逻辑，只关注消息通信
	messageHandlers map[proto.MessageType]conn.MessageHandler
	// session管理器实例
	managerInstance *manager
)

type manager struct {
	mu       sync.RWMutex // 读写锁
	sessions sync.Map     // key: sessionId value: *Session
}

func Init() error {
	messageHandlers = make(map[proto.MessageType]conn.MessageHandler)
	// 注册消息处理器
	registerHandler()
	managerInstance = &manager{}
	return nil
}

func registerHandler() {
	handshakeHandler := handler.NewHandshakeHandler()
	heartbeatHandler := handler.NewHeartbeatHandler()
	disconnectHandler := handler.NewDisconnectHandler()
	messageHandlers[handshakeHandler.Type()] = handshakeHandler
	messageHandlers[heartbeatHandler.Type()] = heartbeatHandler
	messageHandlers[disconnectHandler.Type()] = disconnectHandler
}

// 新建session
func (m *manager) NewSession(conn net.Conn) error {
	session := newSession(conn)
	m.sessions.Store(session.Id, session)
	return nil
}

// 获取session管理器实例
func GetManager() Manager {
	if managerInstance == nil {
		log.Fatal().Msg("session manager not initialized")
	}
	return managerInstance
}

func (m *manager) OpenTunnel(sessionId string, tunnelType proto.TunnelType, targetAddr string) (net.Conn, error) {
	session, ok := m.sessions.Load(sessionId)
	if !ok {
		return nil, errcode.ErrClientDisconnect
	}
	return session.(*Session).openTunnel(tunnelType, targetAddr)
}
