package server

import (
	"fmt"
	"net"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/session"
	"github.com/rs/zerolog/log"
)

type TCPServer struct {
	addr           string          // 监听地址
	ln             net.Listener    // 监听器
	sessionManager session.Manager // 会话管理器
}

// 创建一个TCP服务器
func NewTCPServer() Server {
	cfg := config.Get()
	addr := net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", cfg.Server.TCP.Port))
	return &TCPServer{
		addr:           addr,
		sessionManager: session.GetManager(),
	}
}

// 启动服务器
func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.accept()

	log.Info().Str("addr", s.addr).Msg("tcp server started")
	return nil
}

// 接受连接
func (s *TCPServer) accept() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

// 处理连接
func (s *TCPServer) handleConn(conn net.Conn) {
	if err := s.sessionManager.NewSession(conn); err != nil {
		log.Error().Err(err).Msg("new session error")
		conn.Close()
		return
	}
}

// 停止服务器
func (s *TCPServer) Stop() error {
	if s.ln != nil {
		if err := s.ln.Close(); err != nil {
			return err
		}
	}
	return nil
}
