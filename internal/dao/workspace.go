package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type workspaceDao struct {
	db *gorm.DB
}

func newWorkspaceDao(db *gorm.DB) WorkspaceDao {
	return &workspaceDao{db: db}
}

func (s *workspaceDao) Create(ctx context.Context, name string, description string) (*model.Workspace, error) {
	space := &model.Workspace{
		Name:        name,
		Description: description,
		Status:      1,
	}
	err := s.db.Create(space).Error
	return space, err
}

func (s *workspaceDao) CreateWorkspaceUser(ctx context.Context, workspaceID uint64, userID uint64, role int) error {
	return s.db.Create(&model.WorkspaceUser{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}).Error
}

func (s *workspaceDao) GetByUserID(ctx context.Context, userID uint64) ([]*model.Workspace, error) {
	var list []*model.WorkspaceUser
	err := s.db.Where("user_id = ?", userID).Find(&list).Error
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	var spaces []*model.Workspace
	spaceIds := lo.Map(list, func(item *model.WorkspaceUser, index int) uint64 {
		return item.WorkspaceID
	})
	return spaces, s.db.Where("id IN ?", spaceIds).Find(&spaces).Error
}
