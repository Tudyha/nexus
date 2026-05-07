package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/database"
	"github.com/Tudyha/nexus/internal/model"
	"gorm.io/gorm"
)

type tunnelDao struct {
}

func newTunnelDao(db *gorm.DB) TunnelDao {
	return &tunnelDao{}
}

func (d *tunnelDao) List(ctx context.Context) ([]*model.Tunnel, error) {
	var tunnels []*model.Tunnel
	return tunnels, database.GetDB().Find(&tunnels).Error
}

func (d *tunnelDao) Create(ctx context.Context, tunnel *model.Tunnel) error {
	return database.GetDB().Create(tunnel).Error
}

func (d *tunnelDao) GetByID(ctx context.Context, id uint64) (*model.Tunnel, error) {
	var tunnel model.Tunnel
	err := database.GetDB().First(&tunnel, id).Error
	return &tunnel, err
}

func (d *tunnelDao) ListByClientID(ctx context.Context, clientId uint64) ([]*model.Tunnel, error) {
	var tunnels []*model.Tunnel
	return tunnels, database.GetDB().Where("client_id = ?", clientId).Find(&tunnels).Error
}

func (d *tunnelDao) Delete(ctx context.Context, id uint64) error {
	return database.GetDB().Delete(&model.Tunnel{}, id).Error
}
