package handler

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/conn"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/proto"
)

type DisconnectHandler struct {
	pub           *gochannel.GoChannel
	clientService service.ClientService
}

func NewDisconnectHandler() conn.MessageHandler {
	return &DisconnectHandler{
		pub:           mq.GetPubSub(),
		clientService: service.GetClientService(),
	}
}

func (h *DisconnectHandler) Handle(ctx conn.Context) error {
	sessionId := getSessionId(ctx)

	client, err := h.clientService.GetBySessionID(ctx, sessionId)
	if err != nil {
		return err
	}
	// 更新状态
	if err := h.clientService.UpdateStatus(ctx, client.ID, enum.ClientOffline); err != nil {
		return err
	}

	// 发送mq消息
	if err := h.pub.Publish(constant.MQ_TOPIC_CLIENT_OFFLINE, &message.Message{
		Payload: []byte(sessionId),
	}); err != nil {
		return err
	}
	return nil
}

func (h *DisconnectHandler) Type() proto.MessageType {
	return proto.MessageType_DISCONNECT
}
