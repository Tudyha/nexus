package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/model"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/rs/zerolog/log"
)

// tokenBlacklist 内存 Token 黑名单
var tokenBlacklist struct {
	mu    sync.RWMutex
	token map[string]time.Time // token -> 过期时间
}

func init() {
	tokenBlacklist.token = make(map[string]time.Time)
	// 定期清理过期 token（每小时）
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			tokenBlacklist.mu.Lock()
			now := time.Now()
			for t, expire := range tokenBlacklist.token {
				if now.After(expire) {
					delete(tokenBlacklist.token, t)
				}
			}
			tokenBlacklist.mu.Unlock()
		}
	}()
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func IsTokenBlacklisted(token string) bool {
	tokenBlacklist.mu.RLock()
	defer tokenBlacklist.mu.RUnlock()
	_, ok := tokenBlacklist.token[token]
	return ok
}

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

// Logout 用户登出，将 Token 加入黑名单
func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errcode.ErrInvalidParams
	}
	tokenBlacklist.mu.Lock()
	tokenBlacklist.token[token] = time.Now().Add(24 * time.Hour) // 最多黑名单保留 24h
	tokenBlacklist.mu.Unlock()
	log.Info().Msg("user logged out, token blacklisted")
	return nil
}

// RefreshToken 刷新Token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", errcode.ErrInvalidParams
	}
	return "", errors.New("token 刷新功能待实现")
}

// SetPassword 设置/修改密码
func (s *authService) SetPassword(ctx context.Context, userID uint64, oldPassword, newPassword string) (*model.User, error) {
	user, err := s.userDao.FindByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrUserNotExist
	}

	// 已有密码时需要验证旧密码
	if user.Password != "" {
		if oldPassword == "" {
			return nil, errcode.ErrInvalidParams
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			return nil, errcode.ErrLoginFailed
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hash)

	if err := s.userDao.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
