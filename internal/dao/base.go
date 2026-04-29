package dao

import (
	"context"
	"sync"

	"github.com/Tudyha/nexus/internal/database"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/request"
	"gorm.io/gorm"
)

var (
	once sync.Once

	appDaoInstance       AppDao
	clientDaoInstance    ClientDao
	userDaoInstance      UserDao
	workspaceDaoInstance WorkspaceDao
	tunnelDaoInstance    TunnelDao
)

func Paginate(pageQuery request.PageQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page := pageQuery.Page
		if page <= 0 {
			page = 1
		}

		pageSize := pageQuery.Limit
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

type UserDao interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
}

type WorkspaceDao interface {
	Create(ctx context.Context, name, description string) (*model.Workspace, error)
	CreateWorkspaceUser(ctx context.Context, workspaceID, userID uint64, role int) error
	GetByUserID(ctx context.Context, userID uint64) ([]*model.Workspace, error)
}

type AppDao interface {
	GetByID(ctx context.Context, appID uint64) (*model.App, error)
	Create(ctx context.Context, workspaceID uint64, secret, name, description string) error
	GetAppByWorkspaceIDs(ctx context.Context, workspaceIDs []uint64) ([]*model.App, error)
}

type ClientDao interface {
	Create(ctx context.Context, client *model.Client) error
	GetByDeviceID(ctx context.Context, deviceID string) (*model.Client, error)
	UpdateStatus(ctx context.Context, clientID uint64, status enum.ClientStatus) error
	GetBySessionID(ctx context.Context, sessionID string) (*model.Client, error)
	CreateClientStat(ctx context.Context, stat *model.ClientStat) error
	CountByAppID(ctx context.Context, appID uint64) (online int64, offline int64, err error)
	GetPage(ctx context.Context, appID uint64, query *request.ClientQueryRequest) ([]*model.Client, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.Client, error)
	DeleteByID(ctx context.Context, id uint64) error
	GetByIDs(ctx context.Context, ids []uint64) ([]*model.Client, error)
}

type TunnelDao interface {
	List(ctx context.Context) ([]*model.Tunnel, error)
	Create(ctx context.Context, tunnel *model.Tunnel) error
	ListByClientID(ctx context.Context, clientId uint64) ([]*model.Tunnel, error)
}

func Init() error {
	db := database.GetDB()
	once.Do(func() {
		appDaoInstance = newAppDao(db)
		clientDaoInstance = newClientDao(db)
		userDaoInstance = newUserDao(db)
		workspaceDaoInstance = newWorkspaceDao(db)
		tunnelDaoInstance = newTunnelDao(db)
	})
	return nil
}

func GetAppDao() AppDao {
	return appDaoInstance
}

func GetClientDao() ClientDao {
	return clientDaoInstance
}

func GetUserDao() UserDao {
	return userDaoInstance
}

func GetWorkspaceDao() WorkspaceDao {
	return workspaceDaoInstance
}

func GetTunnelDao() TunnelDao {
	return tunnelDaoInstance
}
