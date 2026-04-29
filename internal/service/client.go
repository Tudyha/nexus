package service

import (
	"context"
	"errors"
	"time"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type clientService struct {
	clientDao dao.ClientDao
}

func newClientService() ClientService {
	return &clientService{
		clientDao: dao.GetClientDao(),
	}
}

func (c *clientService) Connect(ctx context.Context, client *model.Client) error {
	old, err := c.clientDao.GetByDeviceID(ctx, client.DeviceId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	client.ID = old.ID
	client.Status = enum.ClientOnline
	client.LastOnlineTime = time.Now()
	return c.clientDao.Create(ctx, client)
}

func (c *clientService) GetBySessionID(ctx context.Context, sessionID string) (*model.Client, error) {
	return c.clientDao.GetBySessionID(ctx, sessionID)
}

func (c *clientService) UpdateStatus(ctx context.Context, clientID uint64, status enum.ClientStatus) error {
	return c.clientDao.UpdateStatus(ctx, clientID, status)
}

func (c *clientService) CreateClientStat(ctx context.Context, stat *model.ClientStat) error {
	return c.clientDao.CreateClientStat(ctx, stat)
}

func (c *clientService) CountByAppID(ctx context.Context, appID uint64) (online int64, offline int64, err error) {
	return c.clientDao.CountByAppID(ctx, appID)

}

func (c *clientService) GetPage(ctx context.Context, appID uint64, query *request.ClientQueryRequest) (*response.Page[response.ClientResponse], error) {
	clients, total, err := c.clientDao.GetPage(ctx, appID, query)
	if err != nil {
		return nil, err
	}
	var list []response.ClientResponse
	copier.Copy(&list, clients)
	for index := range list {
		list[index].VersionName = "v1.0.0"
	}
	return &response.Page[response.ClientResponse]{
		Total: total,
		List:  list,
	}, nil
}

func (c *clientService) GetByID(ctx context.Context, id uint64) (*model.Client, error) {
	return c.clientDao.GetByID(ctx, id)
}

func (c *clientService) DeleteByID(ctx context.Context, id uint64) (*model.Client, error) {
	old, err := c.clientDao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, errcode.ErrClientNotFound
	}
	return old, c.clientDao.DeleteByID(ctx, id)
}

func (c *clientService) GetByIDs(ctx context.Context, ids []uint64) ([]*model.Client, error) {
	return c.clientDao.GetByIDs(ctx, ids)
}
