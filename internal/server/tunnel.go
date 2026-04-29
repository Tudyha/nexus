package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

type TunnelServer struct {
	tunnelService service.TunnelService

	tunnels []*tunnel
	pub     *gochannel.GoChannel
}

type tunnel struct {
	clientID   uint64 // 客户端id
	network    string // 网络类型 tcp/udp
	localAddr  string // 本地地址
	remoteAddr string // 远程地址
	tunnelType proto.TunnelType

	ln             net.Listener // 监听器
	sessionManager session.Manager
	clientService  service.ClientService
}

func NewTunnelServer() Server {
	s := &TunnelServer{
		tunnelService: service.GetTunnelService(),
		tunnels:       []*tunnel{},
		pub:           mq.GetPubSub(),
	}
	s.initTunnels()
	return s
}

func (s *TunnelServer) Start() error {
	for _, t := range s.tunnels {
		t.start()
	}
	if err := s.subscribe(); err != nil {
		return err
	}
	return nil
}

func (s *TunnelServer) Stop() error {
	for _, t := range s.tunnels {
		t.stop()
	}
	return nil
}

func (s *TunnelServer) subscribe() error {
	// 订阅隧道变更事件
	sub, err := s.pub.Subscribe(context.Background(), constant.MQ_TOPIC_NEW_TUNNEL)
	if err != nil {
		return err
	}
	go func() {
		for msg := range sub {
			var t model.Tunnel
			err := json.Unmarshal(msg.Payload, &t)
			if err != nil {
				continue
			}
			tt := s.registerTunnel(&t)
			tt.start()

			msg.Ack()
		}
	}()
	return nil
}

func (s *TunnelServer) initTunnels() error {
	tunnels, err := s.tunnelService.List(context.Background())
	if err != nil {
		return err
	}
	for _, t := range tunnels {
		s.registerTunnel(t)
	}

	return nil
}

func (s *TunnelServer) registerTunnel(t *model.Tunnel) *tunnel {
	network := ""
	switch t.TunnelType {
	case proto.TunnelType_TCP:
		network = "tcp"
	case proto.TunnelType_UDP:
		network = "udp"
	}

	localAddr := net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", t.LocalPort))
	tt := &tunnel{
		clientID:   t.ClientID,
		network:    network,
		localAddr:  localAddr,
		remoteAddr: t.RemoteAddr,
		tunnelType: t.TunnelType,

		sessionManager: session.GetManager(),
		clientService:  service.GetClientService(),
	}
	s.tunnels = append(s.tunnels, tt)
	return tt
}

func (t *tunnel) start() error {
	ln, err := net.Listen(t.network, t.localAddr)
	if err != nil {
		return err
	}
	t.ln = ln
	go t.accept()

	log.Info().Str("network", t.network).Str("local_addr", t.localAddr).Msg("tunnel listen")
	return nil
}

func (t *tunnel) accept() {
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			return
		}
		go t.handleConn(conn)
	}
}

func (t *tunnel) handleConn(conn net.Conn) {
	var err error
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	client, err := t.clientService.GetByID(context.Background(), t.clientID)
	if err != nil {
		return
	}
	if client.Status != enum.ClientOnline {
		err = errcode.ErrClientNotReady
		return
	}
	target, err := t.sessionManager.OpenTunnel(client.SessionID, t.tunnelType, t.remoteAddr)
	if err != nil {
		return
	}
	go func() {
		defer target.Close()
		defer conn.Close()
		go io.Copy(target, conn)
		io.Copy(conn, target)
	}()
}

func (t *tunnel) stop() error {
	if t.ln != nil {
		return t.ln.Close()
	}
	return nil
}
