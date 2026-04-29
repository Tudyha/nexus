package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/rs/zerolog/log"
)

type smsCode struct {
	code      string
	expiresAt time.Time
}

type smsService struct {
	store sync.Map // key: string(phone+type), value: smsCode
}

func newSmsService() SmsService {
	return &smsService{
		store: sync.Map{},
	}
}

// SendCode 发送验证码
func (s *smsService) SendCode(ctx context.Context, smsCodeType enum.SmsCodeType, target string) error {
	code := utils.GenerateRandomNumber(6)
	log.Info().Str("target", target).Any("smsCodeType", smsCodeType).Str("code", code).Msg("发送验证码")

	// 存储验证码，有效期 5 分钟
	s.store.Store(s.getStoreKey(smsCodeType, target), smsCode{
		code:      code,
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	return nil
}

// VerifyCode 验证验证码
func (s *smsService) VerifyCode(ctx context.Context, smsCodeType enum.SmsCodeType, target string, code string) error {
	key := s.getStoreKey(smsCodeType, target)
	val, ok := s.store.Load(key)
	if !ok {
		return errcode.ErrSmsVerifyCode
	}

	sc := val.(smsCode)
	if time.Now().After(sc.expiresAt) {
		s.store.Delete(key)
		return errcode.ErrSmsVerifyCode
	}

	if sc.code != code {
		return errcode.ErrSmsVerifyCode
	}

	// 验证成功后删除验证码
	s.store.Delete(key)
	return nil
}

func (s *smsService) getStoreKey(smsCodeType enum.SmsCodeType, target string) string {
	return fmt.Sprintf("sms:%d:%s", smsCodeType, target)
}
