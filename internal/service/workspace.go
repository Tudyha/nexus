package service

import (
	"context"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/jinzhu/copier"
)

type workspaceService struct {
	workspaceDao dao.WorkspaceDao
}

func newWorkspaceService() WorkspaceService {
	return &workspaceService{
		workspaceDao: dao.GetWorkspaceDao(),
	}
}

func (s *workspaceService) Create(ctx context.Context, name, description string) (*response.WorkspaceResponse, error) {
	space, err := s.workspaceDao.Create(ctx, name, description)
	if err != nil {
		return nil, err
	}
	var res response.WorkspaceResponse
	copier.Copy(&res, space)
	return &res, nil
}

func (s *workspaceService) GetByUserID(ctx context.Context, userID uint64) ([]*response.WorkspaceResponse, error) {
	list, err := s.workspaceDao.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var res []*response.WorkspaceResponse
	copier.Copy(&res, &list)
	return res, nil
}

func (s *workspaceService) GetByID(ctx context.Context, id uint64) (*response.WorkspaceResponse, error) {
	space, err := s.workspaceDao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var res response.WorkspaceResponse
	copier.Copy(&res, space)
	return &res, nil
}

func (s *workspaceService) List(ctx context.Context) ([]*response.WorkspaceResponse, error) {
	list, err := s.workspaceDao.List(ctx)
	if err != nil {
		return nil, err
	}
	var res []*response.WorkspaceResponse
	copier.Copy(&res, &list)
	return res, nil
}

func (s *workspaceService) Update(ctx context.Context, workspace *model.Workspace) error {
	return s.workspaceDao.Update(ctx, workspace)
}

func (s *workspaceService) Delete(ctx context.Context, id uint64) error {
	return s.workspaceDao.Delete(ctx, id)
}

func (s *workspaceService) ListUsers(ctx context.Context, workspaceID uint64) ([]*response.WorkspaceUserResponse, error) {
	list, err := s.workspaceDao.ListUsers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	var res []*response.WorkspaceUserResponse
	for _, u := range list {
		item := response.WorkspaceUserResponse{
			ID:          u.ID,
			WorkspaceID: u.WorkspaceID,
			UserID:      u.UserID,
			Role:        u.Role,
			Nickname:    u.Nickname,
			Avatar:      u.Avatar,
			CreatedAt:   u.CreatedAt,
		}
		res = append(res, &item)
	}
	return res, nil
}

func (s *workspaceService) AddUser(ctx context.Context, workspaceID, userID uint64, role int) error {
	return s.workspaceDao.CreateWorkspaceUser(ctx, workspaceID, userID, role)
}

func (s *workspaceService) RemoveUser(ctx context.Context, workspaceID, userID uint64) error {
	return s.workspaceDao.DeleteWorkspaceUser(ctx, workspaceID, userID)
}

func (s *workspaceService) UpdateUserRole(ctx context.Context, workspaceID, userID uint64, role int) error {
	return s.workspaceDao.UpdateUserRole(ctx, workspaceID, userID, role)
}
