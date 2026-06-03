package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/enum"
	nexusio "github.com/Tudyha/nexus/pkg/io"
	"github.com/Tudyha/nexus/pkg/metrics"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/rs/zerolog/log"
)

// udpPeer 表示一个 UDP 隧道对端（唯一的源地址对应一个 smux 流）
type udpPeer struct {
	addr   net.Addr
	stream net.Conn
	framer *nexusio.DatagramStream
}

type TunnelServer struct {
	tunnelService service.TunnelService

	tunnels        []*tunnel
	pub            *gochannel.GoChannel
	stopCh         chan struct{}
	healthInterval time.Duration
}

type tunnel struct {
	id         uint64
	clientID   uint64
	network    string // tcp / udp
	localAddr  string
	remoteAddr string
	tunnelType proto.TunnelType

	ln             net.Listener   // TCP 监听器
	packetConn     net.PacketConn // UDP 监听器
	peers          sync.Map       // UDP only: key=addr.String(), value=*udpPeer
	sessionManager session.Manager
	clientService  service.ClientService
	connWG         sync.WaitGroup // 追踪活跃 TCP 连接 / UDP peer 协程
}

func NewTunnelServer() Server {
	s := &TunnelServer{
		tunnelService:  service.GetTunnelService(),
		tunnels:        []*tunnel{},
		pub:            mq.GetPubSub(),
		stopCh:         make(chan struct{}),
		healthInterval: 30 * time.Second,
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
	go s.healthCheck()
	log.Info().Msg("tunnel health check started")
	return nil
}

func (s *TunnelServer) Stop() error {
	close(s.stopCh)

	// 先关闭所有监听器，停止接收新连接
	for _, t := range s.tunnels {
		t.stop()
	}
	// 等待所有活跃隧道连接处理完毕
	for _, t := range s.tunnels {
		t.connWG.Wait()
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
		for {
			select {
			case msg, ok := <-sub:
				if !ok {
					return
				}
				var t model.Tunnel
				err := json.Unmarshal(msg.Payload, &t)
				if err != nil {
					msg.Ack()
					continue
				}
				tt := s.registerTunnel(&t)
				tt.start()
				msg.Ack()
			case <-s.stopCh:
				return
			}
		}
	}()

	// 订阅隧道关闭事件
	closeSub, err := s.pub.Subscribe(context.Background(), constant.MQ_TOPIC_TUNNEL_CLOSE)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg, ok := <-closeSub:
				if !ok {
					return
				}
				tunnelId := utils.StringToUint64(string(msg.Payload))
				if tunnelId != 0 {
					s.deleteTunnel(tunnelId)
				}
				msg.Ack()
			case <-s.stopCh:
				return
			}
		}
	}()

	// 订阅客户端上线事件，自动恢复该客户端的所有隧道
	onlineSub, err := s.pub.Subscribe(context.Background(), constant.MQ_TOPIC_CLIENT_ONLINE)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg, ok := <-onlineSub:
				if !ok {
					return
				}
				sessionId := string(msg.Payload)
				s.rebindClientTunnels(sessionId)
				msg.Ack()
			case <-s.stopCh:
				return
			}
		}
	}()

	// 订阅客户端离线事件，停止该客户端的所有隧道
	offlineSub, err := s.pub.Subscribe(context.Background(), constant.MQ_TOPIC_CLIENT_OFFLINE)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg, ok := <-offlineSub:
				if !ok {
					return
				}
				sessionId := string(msg.Payload)
				s.stopClientTunnels(sessionId)
				msg.Ack()
			case <-s.stopCh:
				return
			}
		}
	}()
	return nil
}

// stopClientTunnels 客户端离线后，停止该客户端的所有隧道监听
func (s *TunnelServer) stopClientTunnels(sessionId string) {
	client, err := service.GetClientService().GetBySessionID(context.Background(), sessionId)
	if err != nil {
		log.Warn().Str("session_id", sessionId).Err(err).Msg("stopClientTunnels: client not found")
		return
	}
	for _, t := range s.tunnels {
		if t.clientID == client.ID && t.isListening() {
			log.Info().Uint64("tunnel_id", t.id).Uint64("client_id", client.ID).
				Msg("stopping tunnel due to client offline")
			t.stop()
		}
	}
}

// healthCheck 定期检查隧道健康状态
func (s *TunnelServer) healthCheck() {
	ticker := time.NewTicker(s.healthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkTunnels()
		case <-s.stopCh:
			return
		}
	}
}

// checkTunnels 检查所有隧道的健康状态，异常时自动恢复
func (s *TunnelServer) checkTunnels() {
	for _, t := range s.tunnels {
		client, err := t.clientService.GetByID(context.Background(), t.clientID)
		if err != nil || client.Status != enum.ClientOnline {
			if t.isListening() {
				log.Warn().Uint64("tunnel_id", t.id).Uint64("client_id", t.clientID).
					Msg("health: stopping tunnel, client offline")
				t.stop()
			}
			continue
		}
		if !t.isListening() {
			log.Warn().Uint64("tunnel_id", t.id).Uint64("client_id", t.clientID).
				Msg("health: restarting stalled tunnel")
			if err := t.start(); err != nil {
				log.Error().Err(err).Uint64("tunnel_id", t.id).Msg("health: restart failed")
			}
		}
	}
}

// isListening 检查隧道是否正在监听
func (t *tunnel) isListening() bool {
	if t.network == "udp" {
		return t.packetConn != nil
	}
	return t.ln != nil
}

// rebindClientTunnels 客户端上线后，检查并恢复其所有隧道
func (s *TunnelServer) rebindClientTunnels(sessionId string) {
	client, err := service.GetClientService().GetBySessionID(context.Background(), sessionId)
	if err != nil {
		log.Warn().Str("session_id", sessionId).Err(err).Msg("rebindTunnels: client not found")
		return
	}

	// 查找该客户端在 DB 中的隧道
	dbTunnels, err := s.tunnelService.ListByClientID(context.Background(), client.ID)
	if err != nil {
		log.Error().Err(err).Uint64("client_id", client.ID).Msg("rebindTunnels: list tunnels failed")
		return
	}

	for _, dt := range dbTunnels {
		// 检查是否已在内存中
		existing := false
		for _, t := range s.tunnels {
			if t.id == dt.ID && t.isListening() {
				existing = true
				break
			}
		}
		if existing {
			continue
		}
		// 注册并启动隧道
		tt := s.registerTunnel(dt)
		if err := tt.start(); err != nil {
			log.Error().Err(err).Uint64("tunnel_id", dt.ID).Msg("rebindTunnels: start failed")
			continue
		}
		log.Info().Uint64("client_id", client.ID).Uint64("tunnel_id", dt.ID).
			Str("addr", tt.localAddr).Msg("tunnel rebound after client reconnect")
	}
}

func (s *TunnelServer) deleteTunnel(tunnelId uint64) {
	for i, t := range s.tunnels {
		if t.id == tunnelId {
			t.stop()
			s.tunnels = append(s.tunnels[:i], s.tunnels[i+1:]...)
			log.Info().Uint64("tunnel_id", tunnelId).Msg("tunnel closed")
			return
		}
	}
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
		id:         t.ID,
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
	if t.network == "udp" {
		pc, err := net.ListenPacket("udp", t.localAddr)
		if err != nil {
			return err
		}
		t.packetConn = pc
		metrics.ActiveTunnels.Inc()
		go t.acceptUDP()
		log.Info().Str("network", "udp").Str("local_addr", t.localAddr).Msg("tunnel listen")
		return nil
	}

	ln, err := net.Listen(t.network, t.localAddr)
	if err != nil {
		return err
	}
	t.ln = ln
	metrics.ActiveTunnels.Inc()
	go t.accept()

	log.Info().Str("network", t.network).Str("local_addr", t.localAddr).Msg("tunnel listen")
	return nil
}

func (t *tunnel) accept() {
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Error().Err(err).Uint64("tunnel_id", t.id).Msg("tunnel accept error")
			return
		}
		go t.handleConn(conn)
	}
}

func (t *tunnel) handleConn(conn net.Conn) {
	t.connWG.Add(1)
	defer t.connWG.Done()
	defer conn.Close()

	// 外网入站连接启用 TCP 优化参数
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(30 * time.Second)
	}
	client, err := t.clientService.GetByID(context.Background(), t.clientID)
	if err != nil {
		return
	}
	if client.Status != enum.ClientOnline {
		return
	}
	s, err := t.sessionManager.GetSession(client.SessionID)
	if err != nil {
		return
	}
	dst, err := s.OpenTunnel(t.tunnelType, t.remoteAddr)
	if err != nil {
		return
	}
	nexusio.Copy(conn, dst)
}

// --- UDP 隧道 ---

// acceptUDP 读取所有 UDP 数据报并分发到对应的 peer
func (t *tunnel) acceptUDP() {
	buf := make([]byte, 65536)
	for {
		n, addr, err := t.packetConn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Error().Err(err).Uint64("tunnel_id", t.id).Msg("udp accept error")
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		t.dispatchUDP(addr, data)
	}
}

// dispatchUDP 将数据报路由到已有 peer，或创建新 peer
func (t *tunnel) dispatchUDP(addr net.Addr, data []byte) {
	key := addr.String()

	// 尝试已有 peer
	if p, ok := t.peers.Load(key); ok {
		peer := p.(*udpPeer)
		if err := peer.framer.WriteDatagram(data); err != nil {
			t.peers.Delete(key)
			peer.stream.Close()
		}
		return
	}

	// 检查客户端状态
	client, err := t.clientService.GetByID(context.Background(), t.clientID)
	if err != nil {
		return
	}
	if client.Status != enum.ClientOnline {
		return
	}
	s, err := t.sessionManager.GetSession(client.SessionID)
	if err != nil {
		return
	}

	// 打开到客户端的 smux 流
	stream, err := s.OpenTunnel(t.tunnelType, t.remoteAddr)
	if err != nil {
		return
	}

	peer := &udpPeer{
		addr:   addr,
		stream: stream,
		framer: nexusio.NewDatagramStream(stream),
	}

	t.peers.Store(key, peer)

	// 发送首个数据报
	if err := peer.framer.WriteDatagram(data); err != nil {
		t.peers.Delete(key)
		stream.Close()
		return
	}

	// 启动响应读取协程：从 smux 流读取帧，写回 UDP 套接字
	t.connWG.Add(1)
	go func() {
		defer t.connWG.Done()
		defer t.peers.Delete(key)
		defer stream.Close()

		for {
			resp, err := peer.framer.ReadDatagram()
			if err != nil {
				return
			}
			if _, err := t.packetConn.WriteTo(resp, addr); err != nil {
				return
			}
		}
	}()
}

func (t *tunnel) stop() error {
	if t.packetConn != nil {
		err := t.packetConn.Close()
		t.packetConn = nil
		metrics.ActiveTunnels.Dec()
		return err
	}
	if t.ln != nil {
		err := t.ln.Close()
		t.ln = nil
		metrics.ActiveTunnels.Dec()
		return err
	}
	return nil
}
