package dao

import (
	"context"
	"time"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/request"
	"gorm.io/gorm"
)

type clientDao struct {
	db *gorm.DB
}

func newClientDao(db *gorm.DB) ClientDao {
	return &clientDao{
		db: db,
	}
}

func (c *clientDao) Create(ctx context.Context, client *model.Client) error {
	return c.db.WithContext(ctx).Save(client).Error
}

func (c *clientDao) GetByDeviceID(ctx context.Context, deviceID string) (*model.Client, error) {
	var client model.Client
	return &client, c.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&client).Error
}

func (c *clientDao) UpdateStatus(ctx context.Context, clientID uint64, status enum.ClientStatus) error {
	return c.db.WithContext(ctx).Model(&model.Client{}).Where("id = ?", clientID).Update("status", status).Update("last_online_time", time.Now()).Error
}

func (c *clientDao) GetBySessionID(ctx context.Context, sessionID string) (*model.Client, error) {
	var client model.Client
	return &client, c.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&client).Error
}

func (c *clientDao) CreateClientStat(ctx context.Context, stat *model.ClientStat) error {
	return c.db.WithContext(ctx).Create(stat).Error
}

func (c *clientDao) CountByAppID(ctx context.Context, appID uint64) (online int64, offline int64, err error) {
	err = c.db.WithContext(ctx).Model(&model.Client{}).Where("app_id = ? AND status = ?", appID, enum.ClientOnline).Count(&online).Error
	if err != nil {
		return
	}
	err = c.db.WithContext(ctx).Model(&model.Client{}).Where("app_id = ? AND status = ?", appID, enum.ClientOffline).Count(&offline).Error
	return
}

func (c *clientDao) GetPage(ctx context.Context, appID uint64, query *request.ClientQueryRequest) ([]*model.Client, int64, error) {
	var clients []*model.Client
	var total int64
	db := c.db.WithContext(ctx).Model(&model.Client{})
	if appID != 0 {
		db = db.Where("app_id = ?", appID)
	}
	if query.Status != 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.Hostname != "" {
		db = db.Where("hostname like ?", "%"+query.Hostname+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db = db.Order("id desc")
	return clients, total, db.Scopes(Paginate(query.PageQuery)).Find(&clients).Error
}

func (c *clientDao) GetByID(ctx context.Context, id uint64) (*model.Client, error) {
	var client model.Client
	return &client, c.db.WithContext(ctx).Where("id = ?", id).First(&client).Error
}

func (c *clientDao) DeleteByID(ctx context.Context, id uint64) error {
	return c.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Client{}).Error
}

func (c *clientDao) GetByIDs(ctx context.Context, ids []uint64) ([]*model.Client, error) {
	var clients []*model.Client
	return clients, c.db.WithContext(ctx).Where("id in ?", ids).Find(&clients).Error
}

func (c *clientDao) ListOnlineByAppID(ctx context.Context, appID uint64) ([]*model.Client, error) {
	var clients []*model.Client
	return clients, c.db.WithContext(ctx).
		Where("app_id = ? AND status = ?", appID, enum.ClientOnline).
		Order("id desc").
		Find(&clients).Error
}
