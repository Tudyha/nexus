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

	appDaoInstance           AppDao
	clientDaoInstance        ClientDao
	userDaoInstance          UserDao
	workspaceDaoInstance     WorkspaceDao
	tunnelDaoInstance        TunnelDao
	versionDaoInstance       VersionDao
	taskDaoInstance          TaskDao
	taskExecutionDaoInstance TaskExecutionDao
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
	List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
}

type WorkspaceDao interface {
	Create(ctx context.Context, name, description string) (*model.Workspace, error)
	CreateWorkspaceUser(ctx context.Context, workspaceID, userID uint64, role int) error
	GetByUserID(ctx context.Context, userID uint64) ([]*model.Workspace, error)
	GetByID(ctx context.Context, id uint64) (*model.Workspace, error)
	List(ctx context.Context) ([]*model.Workspace, error)
	Update(ctx context.Context, workspace *model.Workspace) error
	Delete(ctx context.Context, id uint64) error
	ListUsers(ctx context.Context, workspaceID uint64) ([]*WorkspaceUserInfo, error)
	DeleteWorkspaceUser(ctx context.Context, workspaceID, userID uint64) error
	UpdateUserRole(ctx context.Context, workspaceID, userID uint64, role int) error
}

type AppDao interface {
	GetByID(ctx context.Context, appID uint64) (*model.App, error)
	Create(ctx context.Context, workspaceID uint64, secret, name, description string) error
	GetAppByWorkspaceIDs(ctx context.Context, workspaceIDs []uint64) ([]*model.App, error)
	UpdateConfig(ctx context.Context, appID uint64, config string) error
	ListByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.App, error)
	Update(ctx context.Context, app *model.App) error
	Delete(ctx context.Context, id uint64) error
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
	ListOnlineByAppID(ctx context.Context, appID uint64) ([]*model.Client, error)
}

type TunnelDao interface {
	List(ctx context.Context) ([]*model.Tunnel, error)
	Create(ctx context.Context, tunnel *model.Tunnel) error
	GetByID(ctx context.Context, id uint64) (*model.Tunnel, error)
	ListByClientID(ctx context.Context, clientId uint64) ([]*model.Tunnel, error)
	Delete(ctx context.Context, id uint64) error
}

type VersionDao interface {
	Create(ctx context.Context, version *model.Version) error
	GetByID(ctx context.Context, id uint64) (*model.Version, error)
	GetLatestByOS(ctx context.Context, os, arch string) (*model.Version, error)
	GetPage(ctx context.Context, query request.PageQuery) ([]*model.Version, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type TaskDao interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uint64) (*model.Task, error)
}

type TaskExecutionDao interface {
	Create(ctx context.Context, execs []*model.TaskExecution) error
	Update(ctx context.Context, exec *model.TaskExecution) error
	GetByID(ctx context.Context, id uint64) (*model.TaskExecution, error)
	GetLatestByClientID(ctx context.Context, clientID uint64) (*model.TaskExecution, error)
	GetLatestByClientIDs(ctx context.Context, clientIDs []uint64) (map[uint64]*model.TaskExecution, error)
}

func Init() error {
	db := database.GetDB()
	once.Do(func() {
		appDaoInstance = newAppDao(db)
		clientDaoInstance = newClientDao(db)
		userDaoInstance = newUserDao(db)
		workspaceDaoInstance = newWorkspaceDao(db)
		tunnelDaoInstance = newTunnelDao(db)
		versionDaoInstance = newVersionDao(db)
		taskDaoInstance = newTaskDao(db)
		taskExecutionDaoInstance = newTaskExecutionDao(db)
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

func GetVersionDao() VersionDao {
	return versionDaoInstance
}

func GetTaskDao() TaskDao {
	return taskDaoInstance
}

func GetTaskExecutionDao() TaskExecutionDao {
	return taskExecutionDaoInstance
}
