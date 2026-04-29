package handler

import (
	"net"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/conn"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/ip"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"
)

type HandshakeHandler struct {
	appService    service.AppService
	clientService service.ClientService
	pub           *gochannel.GoChannel
}

func NewHandshakeHandler() conn.MessageHandler {
	return &HandshakeHandler{
		appService:    service.GetAppService(),
		clientService: service.GetClientService(),
		pub:           mq.GetPubSub(),
	}
}

func (h *HandshakeHandler) Handle(ctx conn.Context) error {
	var req proto.HandshakeReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	log.Info().Any("req", &req).Msg("handshake req")

	// 获取应用信息
	app, err := h.appService.GetApp(ctx, uint64(req.AppId))
	if err != nil {
		return err
	}
	if app == nil {
		return errcode.ErrClientAuthFailed
	}
	// 验证签名
	if err := utils.VerifySignature(app.ID, app.AppSecret, req.Timestamp, req.Nonce, req.Signature); err != nil {
		return errcode.ErrClientAuthFailed
	}

	// 保存客户端信息
	clientInfo := req.ClientInfo
	var client model.Client
	if err := copier.Copy(&client, clientInfo); err != nil {
		log.Error().Err(err).Msg("copy client info failed")
		return errcode.ErrInternalServer
	}

	sessionId := getSessionId(ctx)

	client.Version = uint32(req.Version)
	client.SessionID = sessionId
	client.AppID = uint64(req.AppId)
	client.Version = uint32(req.Version)

	remoteAddr := ctx.GetConn().RemoteAddr()
	remoteIP, port, err := net.SplitHostPort(remoteAddr.String())
	if err != nil {
		return err
	}
	client.RemoteIP = remoteIP
	client.Port = port
	client.RemoteIpCountry = ip.GetIPCountry(client.RemoteIP)
	if err := h.clientService.Connect(ctx, &client); err != nil {
		return err
	}

	h.pub.Publish(constant.MQ_TOPIC_CLIENT_ONLINE, &message.Message{
		Payload: []byte(sessionId),
	})

	return nil
}

func (h *HandshakeHandler) Type() proto.MessageType {
	return proto.MessageType_HANDSHAKE
}
