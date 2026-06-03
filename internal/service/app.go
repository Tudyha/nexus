package service

import (
	"context"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type appService struct {
	appDao dao.AppDao
}

func newAppService() AppService {
	return &appService{
		appDao: dao.GetAppDao(),
	}
}

func (a *appService) GetApp(ctx context.Context, appID uint64) (*model.App, error) {
	return a.appDao.GetByID(ctx, appID)
}

func (a *appService) UpdateConfig(ctx context.Context, appID uint64, config string) error {
	return a.appDao.UpdateConfig(ctx, appID, config)
}

func (s *appService) GetByID(ctx context.Context, appID uint64) (*response.AppResponse, error) {
	app, err := s.appDao.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	var res response.AppResponse
	copier.Copy(&res, app)
	return &res, nil
}

func (s *appService) ListByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*response.AppResponse, error) {
	list, err := s.appDao.ListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	var res []*response.AppResponse
	copier.Copy(&res, &list)
	return res, nil
}

func (s *appService) Create(ctx context.Context, workspaceID uint64, name, description string) (*response.AppResponse, error) {
	secret := uuid.New().String()
	app := &model.App{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		AppSecret:   secret,
		Status:      1,
	}
	if err := s.appDao.Create(ctx, workspaceID, secret, name, description); err != nil {
		return nil, err
	}
	var res response.AppResponse
	copier.Copy(&res, app)
	return &res, nil
}

func (s *appService) Update(ctx context.Context, app *model.App) error {
	return s.appDao.Update(ctx, app)
}

func (s *appService) Delete(ctx context.Context, id uint64) error {
	return s.appDao.Delete(ctx, id)
}
