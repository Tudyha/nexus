package dao

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// WorkspaceUserInfo holds workspace user with joined user info
type WorkspaceUserInfo struct {
	model.WorkspaceUser
	Nickname string
	Avatar   string
}

type workspaceDao struct {
	db *gorm.DB
}

func newWorkspaceDao(db *gorm.DB) WorkspaceDao {
	return &workspaceDao{db: db}
}

func (s *workspaceDao) GetByID(ctx context.Context, id uint64) (*model.Workspace, error) {
	var space model.Workspace
	err := s.db.WithContext(ctx).First(&space, id).Error
	if err != nil {
		return nil, err
	}
	return &space, nil
}

func (s *workspaceDao) List(ctx context.Context) ([]*model.Workspace, error) {
	var list []*model.Workspace
	return list, s.db.WithContext(ctx).Find(&list).Error
}

func (s *workspaceDao) Update(ctx context.Context, workspace *model.Workspace) error {
	return s.db.WithContext(ctx).Model(&model.Workspace{}).Where("id = ?", workspace.ID).Updates(workspace).Error
}

func (s *workspaceDao) Delete(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Delete(&model.Workspace{}, id).Error
}

func (s *workspaceDao) ListUsers(ctx context.Context, workspaceID uint64) ([]*WorkspaceUserInfo, error) {
	var list []*WorkspaceUserInfo
	err := s.db.WithContext(ctx).
		Table("t_workspace_user").
		Select("t_workspace_user.*, t_user.nickname, t_user.avatar").
		Joins("left join t_user on t_user.id = t_workspace_user.user_id").
		Where("t_workspace_user.workspace_id = ?", workspaceID).
		Find(&list).Error
	return list, err
}

func (s *workspaceDao) DeleteWorkspaceUser(ctx context.Context, workspaceID, userID uint64) error {
	return s.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&model.WorkspaceUser{}).Error
}

func (s *workspaceDao) UpdateUserRole(ctx context.Context, workspaceID, userID uint64, role int) error {
	return s.db.WithContext(ctx).
		Model(&model.WorkspaceUser{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", role).Error
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
