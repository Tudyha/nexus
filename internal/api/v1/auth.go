package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	smsService  service.SmsService
	authService service.AuthService
}

func newAuthController() *AuthController {
	return &AuthController{
		smsService:  service.GetSmsService(),
		authService: service.GetAuthService(),
	}
}

// 登出
func (h *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token != "" {
		h.authService.Logout(ctx, token)
	}
	response.Success(ctx, nil)
}

// 刷新 Token
func (h *AuthController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	token, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, token)
}

// 设置密码
func (h *AuthController) SetPassword(ctx *gin.Context) {
	var req request.SetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	userId := getUserId(ctx)
	user, err := h.authService.SetPassword(ctx, userId, req.Password, req.NewPassword)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, user)
}

// 发送验证码
func (h *AuthController) SendCode(ctx *gin.Context) {
	var req request.SmsSendCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}

	if err := h.smsService.SendCode(ctx, req.Type, req.Target); err != nil {
		response.FailWithMsg(ctx, errcode.ErrSmsSendCode, err.Error())
		return
	}
	response.Success(ctx, nil)
}
