package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/request"
	"gorm.io/gorm"
)

type versionDao struct {
	db *gorm.DB
}

func newVersionDao(db *gorm.DB) VersionDao {
	return &versionDao{
		db: db,
	}
}

func (d *versionDao) Create(ctx context.Context, version *model.Version) error {
	return d.db.WithContext(ctx).Create(version).Error
}

func (d *versionDao) GetByID(ctx context.Context, id uint64) (*model.Version, error) {
	var version model.Version
	return &version, d.db.WithContext(ctx).Where("id = ?", id).First(&version).Error
}

func (d *versionDao) GetLatestByOS(ctx context.Context, os, arch string) (*model.Version, error) {
	var version model.Version
	return &version, d.db.WithContext(ctx).Where("os = ? AND arch = ?", os, arch).Order("version desc").First(&version).Error
}

func (d *versionDao) GetPage(ctx context.Context, query request.PageQuery) ([]*model.Version, int64, error) {
	var versions []*model.Version
	var total int64
	db := d.db.WithContext(ctx).Model(&model.Version{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db = db.Order("id desc")
	return versions, total, db.Scopes(Paginate(query)).Find(&versions).Error
}

func (d *versionDao) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Version{}).Error
}
