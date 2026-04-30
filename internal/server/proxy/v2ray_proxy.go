package proxy

import (
	"context"
	"fmt"

	"github.com/Tudyha/nexus/internal/session"
	"github.com/Tudyha/nexus/pkg/proto"
	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	v2raySession "github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		h := new(handler)
		if err := core.RequireFeatures(ctx, func() error {
			return h.Init(config.(*SimplifiedConfig))
		}); err != nil {
			return nil, err
		}
		return h, nil
	}))
}

type handler struct {
	config         *SimplifiedConfig
	sessionManager session.Manager
}

func (h *handler) Init(config *SimplifiedConfig) error {
	h.config = config
	h.sessionManager = session.GetManager()
	return nil
}

func (h *handler) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := v2raySession.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return fmt.Errorf("target not specified")
	}
	destination := outbound.Target

	input := link.Reader
	output := link.Writer

	account := v2raySession.InboundFromContext(ctx).User.Account.(*vmess.MemoryAccount)

	// 获取客户端session
	s, err := h.sessionManager.GetSession(account.ID.String())
	if err != nil {
		return err
	}

	var tunnelType proto.TunnelType
	switch destination.Network {
	case net.Network_TCP:
		tunnelType = proto.TunnelType_TCP
	case net.Network_UDP:
		tunnelType = proto.TunnelType_UDP
	default:
		return fmt.Errorf("unknown network: %s", destination.Network)
	}

	// 打开隧道
	src, err := s.OpenTunnel(tunnelType, destination.NetAddr())
	if err != nil {
		return err
	}

	requestDone := func() error {
		var writer buf.Writer
		if destination.Network == net.Network_TCP {
			writer = buf.NewWriter(src)
		} else {
			writer = &buf.SequentialWriter{Writer: src}
		}

		if err := buf.Copy(input, writer); err != nil {
			return err
		}
		// 请求发完，关闭写端通知对端没有更多数据
		// 不关闭整个 src，让 responseDone 继续读
		if tc, ok := src.(interface{ CloseWrite() error }); ok {
			tc.CloseWrite()
		}
		return nil
	}

	responseDone := func() error {
		var reader buf.Reader
		if destination.Network == net.Network_TCP {
			reader = buf.NewReader(src)
		} else {
			reader = buf.NewPacketReader(src)
		}
		return buf.Copy(reader, output)
	}

	err = task.Run(ctx, requestDone, responseDone)
	src.Close()
	return err
}
