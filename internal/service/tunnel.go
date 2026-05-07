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
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/utils"
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

func (s *tunnelService) Delete(ctx context.Context, clientId uint64, tunnelId uint64) error {
	tunnel, err := s.tunnelDao.GetByID(ctx, tunnelId)
	if err != nil {
		return errcode.ErrNotFound
	}
	if tunnel.ClientID != clientId {
		return errcode.ErrNotFound
	}
	if err := s.tunnelDao.Delete(ctx, tunnelId); err != nil {
		return err
	}
	return s.pub.Publish(constant.MQ_TOPIC_TUNNEL_CLOSE, &message.Message{
		Payload: []byte(utils.Uint64ToString(tunnelId)),
	})
}
