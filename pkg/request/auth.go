package request

import "github.com/Tudyha/nexus/pkg/enum"

// 发送验证码请求
type SmsSendCodeRequest struct {
	Target string           `json:"target"  binding:"required"` // 手机号或邮箱
	Type   enum.SmsCodeType `json:"type" binding:"required"`    // 1:登录
}

// 设置密码请求
type SetPasswordRequest struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// 登录请求
type LoginRequest struct {
	LoginType enum.LoginType `json:"login_type" binding:"required,oneof=1 2"` // 1:验证码登录 2:密码登录
	Username  string         `json:"username" binding:"required"`             // 用户名、手机号或邮箱
	Password  string         `json:"password"`                                // 密码登录时必填
	Code      string         `json:"code"`                                    // 验证码登录时必填
}
