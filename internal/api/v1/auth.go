package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
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

// SendCode 发送验证码
// @Summary 发送验证码
// @Tags Auth
// @Accept json
// @Produce json
// @Param smsSendCodeRequest body request.SmsSendCodeRequest true "登录请求"
// @Success 200 {object} response.Response "成功响应"
// @Router /auth/send_code [post]
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

// Login 用户登录
// @Summary 用户登录
// @Description 支持手机验证码、密码登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body request.LoginRequest true "登录请求"
// @Success 200 {object} response.LoginResponse "成功响应"
// @Router /auth/login [post]
func (h *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	res, err := h.authService.Login(ctx, &req)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, res)
}
