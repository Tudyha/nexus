package service

import (
	"context"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
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
