package service

import (
	"context"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
)

var (
	appServiceInstance    AppService
	clientServiceInstance ClientService
	smsServiceInstance    SmsService
	userServiceInstance   UserService
	authServiceInstance   AuthService
	tunnelServiceInstance TunnelService
)

// AppService 应用服务接口
type AppService interface {
	GetApp(ctx context.Context, appId uint64) (*model.App, error)
}

// ClientService 客户端服务接口
type ClientService interface {
	Connect(ctx context.Context, client *model.Client) error
	GetBySessionID(ctx context.Context, sessionID string) (*model.Client, error)
	UpdateStatus(ctx context.Context, clientID uint64, status enum.ClientStatus) error
	CreateClientStat(ctx context.Context, stat *model.ClientStat) error
	CountByAppID(ctx context.Context, appID uint64) (online int64, offline int64, err error)
	GetPage(ctx context.Context, appID uint64, query *request.ClientQueryRequest) (*response.Page[response.ClientResponse], error)
	GetByID(ctx context.Context, id uint64) (*model.Client, error)
	DeleteByID(ctx context.Context, id uint64) (*model.Client, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]*model.Client, error)
}

// SmsService 短信服务接口
type SmsService interface {
	SendCode(ctx context.Context, smsCodeType enum.SmsCodeType, phone string) error
	VerifyCode(ctx context.Context, smsCodeType enum.SmsCodeType, phone, code string) error
}

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, req *request.LoginRequest) (*model.User, error)
	SendCode(ctx context.Context, req *request.SmsSendCodeRequest) error
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
}

// UserService 用户服务接口
type UserService interface {
	Create(ctx context.Context, user *model.User) error
	Detail(ctx context.Context, id uint64) (*response.UserResponse, error)
}

// TunnelService 隧道服务接口
type TunnelService interface {
	List(ctx context.Context) ([]*model.Tunnel, error)
	Create(ctx context.Context, clientId uint64, tunnel *request.TunnelCreateRequest) error
	ListByClientID(ctx context.Context, clientId uint64) ([]*model.Tunnel, error)
	Delete(ctx context.Context, clientId uint64, tunnelId uint64) error
}

func Init() error {
	appServiceInstance = newAppService()
	clientServiceInstance = newClientService()
	smsServiceInstance = newSmsService()
	userServiceInstance = newUserService()
	authServiceInstance = newAuthService(smsServiceInstance, userServiceInstance)
	tunnelServiceInstance = newTunnelService()
	return nil
}

func GetAppService() AppService {
	return appServiceInstance
}

func GetClientService() ClientService {
	return clientServiceInstance
}

func GetSmsService() SmsService {
	return smsServiceInstance
}

func GetUserService() UserService {
	return userServiceInstance
}

func GetAuthService() AuthService {
	return authServiceInstance
}

func GetTunnelService() TunnelService {
	return tunnelServiceInstance
}
