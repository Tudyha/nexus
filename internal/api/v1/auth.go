package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	smsService service.SmsService
}

func newAuthController() *AuthController {
	return &AuthController{
		smsService: service.GetSmsService(),
	}
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
