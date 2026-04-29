package service

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/mq"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/request"
)

type tunnelService struct {
	tunnelDao dao.TunnelDao
	pub       *gochannel.GoChannel
}

func newTunnelService() *tunnelService {
	return &tunnelService{
		tunnelDao: dao.GetTunnelDao(),
		pub:       mq.GetPubSub(),
	}
}

func (s *tunnelService) List(ctx context.Context) ([]*model.Tunnel, error) {
	return s.tunnelDao.List(ctx)
}

func (s *tunnelService) Create(ctx context.Context, clientId uint64, req *request.TunnelCreateRequest) error {
	tunnel := &model.Tunnel{
		ClientID:   clientId,
		TunnelType: req.TunnelType,
		LocalPort:  req.LocalPort,
		RemoteAddr: req.RemoteAddr,
	}
	if err := s.tunnelDao.Create(ctx, tunnel); err != nil {
		return err
	}
	bytes, err := json.Marshal(tunnel)
	if err != nil {
		return err
	}
	return s.pub.Publish(constant.MQ_TOPIC_NEW_TUNNEL, &message.Message{
		Payload: bytes,
	})
}

func (s *tunnelService) ListByClientID(ctx context.Context, clientId uint64) ([]*model.Tunnel, error) {
	return s.tunnelDao.ListByClientID(ctx, clientId)
}
