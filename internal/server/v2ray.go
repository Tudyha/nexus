package server

import (
	"context"
	"fmt"

	gonet "net"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/server/proxy"
	_ "github.com/Tudyha/nexus/internal/server/proxy"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/utils"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	proxyInbound "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"
	"github.com/v2fly/v2ray-core/v5/app/router"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/uuid"
	"github.com/v2fly/v2ray-core/v5/features/inbound"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	vmessInbound "github.com/v2fly/v2ray-core/v5/proxy/vmess/inbound"
	"google.golang.org/protobuf/types/known/anypb"

	zeroLog "github.com/rs/zerolog/log"
)

const (
	inbound_tag = "vmess-in"
)

type v2rayServer struct {
	addr     string
	logpath  string
	config   *core.Config   // v2ray config
	instance *core.Instance // v2ray instance
	sub      *gochannel.GoChannel
	stopCh   chan struct{}
}

func NewV2rayServer() Server {
	cfg := config.Get()
	s := &v2rayServer{}

	s.addr = fmt.Sprintf("%s:%d", "0.0.0.0", cfg.Server.V2ray.Port)
	s.logpath = cfg.Server.V2ray.LogPath
	s.config = initConfig(s.addr, s.logpath)
	s.sub = mq.GetPubSub()
	s.stopCh = make(chan struct{})

	return s
}

// initConfig 初始化v2ray配置
func initConfig(addr string, logpath string) *core.Config {
	host, port, _ := gonet.SplitHostPort(addr)
	p := utils.StringToUint64(port)

	return &core.Config{
		// app配置
		App: []*anypb.Any{
			serial.ToTypedMessage(&log.Config{
				Access: &log.LogSpecification{
					Type:  log.LogType_File,
					Level: 1,
					Path:  logpath,
				},
				Error: &log.LogSpecification{
					Type:  log.LogType_File,
					Level: 1,
					Path:  logpath,
				},
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(&router.Config{}),
		},

		// 入口配置
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(net.Port(p)),
					Listen:    net.NewIPOrDomain(net.ParseAddress(host)),
				}),
				ProxySettings: serial.ToTypedMessage(&vmessInbound.SimplifiedConfig{
					Users: []string{},
				}),
				Tag: inbound_tag,
			},
		},

		// 出口配置
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&proxy.SimplifiedConfig{}),
			},
		},
	}
}

func (v *v2rayServer) Start() error {
	instance, err := core.New(v.config)
	if err != nil {
		return err
	}

	v.instance = instance
	v.subscribe()
	if err := v.instance.Start(); err != nil {
		return err
	}
	zeroLog.Info().Str("addr", v.addr).Msg("v2ray started")
	return nil
}

func (v *v2rayServer) Stop() error {
	close(v.stopCh)
	if v.instance != nil {
		if err := v.instance.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (v *v2rayServer) subscribe() {
	go v.subscribeClientOnline()
	go v.subscribeClientOffline()
}

func (v *v2rayServer) subscribeClientOnline() {
	messages, err := v.sub.Subscribe(context.Background(), constant.MQ_TOPIC_CLIENT_ONLINE)
	if err != nil {
		zeroLog.Error().Err(err).Msg("v2ray: subscribe client online failed")
		return
	}
	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			clientSessionId := string(msg.Payload)
			zeroLog.Info().Str("clientSessionId", clientSessionId).Msg("v2ray receive mq client online msg")
			if err := v.addVMessUser(clientSessionId); err != nil {
				zeroLog.Error().Err(err).Str("sessionId", clientSessionId).Msg("v2ray: add vmess user failed")
			}
			msg.Ack()
		case <-v.stopCh:
			return
		}
	}
}

func (v *v2rayServer) subscribeClientOffline() {
	messages, err := v.sub.Subscribe(context.Background(), constant.MQ_TOPIC_CLIENT_OFFLINE)
	if err != nil {
		zeroLog.Error().Err(err).Msg("v2ray: subscribe client offline failed")
		return
	}
	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			clientSessionId := string(msg.Payload)
			zeroLog.Info().Str("clientSessionId", clientSessionId).Msg("v2ray receive mq client offline msg")

			if err := v.removeVMessUser(clientSessionId); err != nil {
				zeroLog.Error().Err(err).Str("sessionId", clientSessionId).Msg("v2ray: remove vmess user failed")
			}

			msg.Ack()
		case <-v.stopCh:
			return
		}
	}
}

func (v *v2rayServer) addVMessUser(id string) error {
	h, err := v.getInboundHandler()
	if err != nil {
		return err
	}
	uid, _ := uuid.ParseString(id)
	return h.AddUser(context.Background(), &protocol.MemoryUser{
		Account: &vmess.MemoryAccount{
			ID: protocol.NewID(uid),
		},
		Email: id,
	})
}

func (v *v2rayServer) removeVMessUser(id string) error {
	h, err := v.getInboundHandler()
	if err != nil {
		return err
	}
	return h.RemoveUser(context.Background(), id)
}

func (v *v2rayServer) getInboundHandler() (*vmessInbound.Handler, error) {
	handlerMgr := v.instance.GetFeature(inbound.ManagerType()).(inbound.Manager)

	handler, err := handlerMgr.GetHandler(context.Background(), inbound_tag)
	if err != nil {
		return nil, err
	}
	h, ok := handler.(*proxyInbound.AlwaysOnInboundHandler)
	if !ok {
		return nil, fmt.Errorf("handler is not AlwaysOnInboundHandler")
	}
	i, ok := h.GetInbound().(*vmessInbound.Handler)
	if !ok {
		return nil, fmt.Errorf("inbound is not vmessInbound.Handler")
	}
	return i, nil
}
