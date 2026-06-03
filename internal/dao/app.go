package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"gorm.io/gorm"
)

type appDao struct {
	db *gorm.DB
}

func newAppDao(db *gorm.DB) AppDao {
	return &appDao{db: db}
}

func (a *appDao) ListByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.App, error) {
	var list []*model.App
	return list, a.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&list).Error
}

func (a *appDao) Update(ctx context.Context, app *model.App) error {
	return a.db.WithContext(ctx).Model(&model.App{}).Where("id = ?", app.ID).Updates(app).Error
}

func (a *appDao) Delete(ctx context.Context, id uint64) error {
	return a.db.WithContext(ctx).Delete(&model.App{}, id).Error
}

func (a *appDao) GetByID(ctx context.Context, appID uint64) (*model.App, error) {
	var app model.App
	return &app, a.db.WithContext(ctx).First(&app, appID).Error
}

func (a *appDao) Create(ctx context.Context, workspaceID uint64, secret string, name string, description string) error {
	return a.db.WithContext(ctx).Create(&model.App{
		WorkspaceID: workspaceID,
		AppSecret:   secret,
		Name:        name,
		Description: description,
		Status:      1,
	}).Error
}

func (a *appDao) GetAppByWorkspaceIDs(ctx context.Context, workspaceIDs []uint64) ([]*model.App, error) {
	var list []*model.App
	return list, a.db.WithContext(ctx).Where("workspace_id IN ?", workspaceIDs).Find(&list).Error
}

func (a *appDao) UpdateConfig(ctx context.Context, appID uint64, config string) error {
	return a.db.WithContext(ctx).Model(&model.App{}).Where("id = ?", appID).Update("config", config).Error
}
