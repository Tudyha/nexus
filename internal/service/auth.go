package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
)

type authService struct {
	userDao     dao.UserDao
	smsService  SmsService
	userService UserService
}

func newAuthService(smsService SmsService, userService UserService) AuthService {
	return &authService{
		userDao:     dao.GetUserDao(),
		smsService:  smsService,
		userService: userService,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*model.User, error) {
	var user *model.User
	var err error

	// 根据登录类型处理登录逻辑
	switch req.LoginType {
	case enum.LoginTypeCode:
		user, err = s.loginByCode(ctx, req.Username, req.Code)
		if err != nil {
			return nil, err
		}
	case enum.LoginTypePassword:
		user, err = s.loginByPassword(ctx, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errcode.ErrLoginFailed
	}
	return user, nil
}

// loginByPassword 密码登录
func (s *authService) loginByPassword(ctx context.Context, identifier, password string) (*model.User, error) {
	var user *model.User
	var err error

	user, err = s.userDao.FindByUsername(ctx, identifier)

	if err != nil {
		return nil, errcode.ErrLoginFailed
	}

	if err := s.checkUserStatus(user); err != nil {
		return nil, err
	}
	if user.Password == "" {
		return nil, errcode.ErrAuthNotSetPassword
	}
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errcode.ErrLoginFailed
	}
	return user, nil
}

// loginByCode 验证码登录
func (s *authService) loginByCode(ctx context.Context, phone, code string) (*model.User, error) {
	if err := s.smsService.VerifyCode(ctx, enum.SmsCodeTypeLogin, phone, code); err != nil {
		return nil, err
	}
	user, err := s.userDao.FindByUsername(ctx, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，注册新用户
			user, err = s.registerByPhone(ctx, phone)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if err := s.checkUserStatus(user); err != nil {
		return nil, err
	}
	return user, nil
}

// SendCode 发送验证码
func (s *authService) SendCode(ctx context.Context, req *request.SmsSendCodeRequest) error {
	return s.smsService.SendCode(ctx, req.Type, req.Target)
}

// checkUserStatus 检查用户状态
func (s *authService) checkUserStatus(user *model.User) error {
	if user.Status != 1 {
		return errcode.ErrUserDisabled
	}
	return nil
}

// register 手机号注册新用户
func (s *authService) registerByPhone(ctx context.Context, phone string) (*model.User, error) {
	user := &model.User{
		Username: phone,
		Phone:    phone,
		Status:   1,
		Avatar:   constant.DefaultAvatar,
	}
	if err := s.userService.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, token string) error {
	// TODO: 实现Token黑名单逻辑
	// 可以将Token添加到Redis的黑名单中，设置过期时间
	return nil
}

// RefreshToken 刷新Token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// TODO: 实现Token刷新逻辑
	// 验证refreshToken，生成新的accessToken
	return "", errors.New("Token刷新功能待实现")
}
