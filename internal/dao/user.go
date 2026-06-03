package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"gorm.io/gorm"
)

type userDao struct {
	db *gorm.DB
}

func newUserDao(db *gorm.DB) UserDao {
	return &userDao{db: db}
}

// Create 创建用户
func (d *userDao) Create(ctx context.Context, user *model.User) error {
	return d.db.Create(user).Error
}

// Update 更新用户
func (d *userDao) Update(ctx context.Context, user *model.User) error {
	return d.db.Save(user).Error
}

func (d *userDao) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := d.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (d *userDao) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := d.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *userDao) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	d.db.Model(&model.User{}).Count(&total)
	offset := (page - 1) * pageSize
	err := d.db.Offset(offset).Limit(pageSize).Order("id desc").Find(&users).Error
	return users, total, err
}
