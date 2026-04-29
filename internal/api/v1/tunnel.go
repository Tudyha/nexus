package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type TunnelController struct {
	tunnelService service.TunnelService
}

func newTunnelController() *TunnelController {
	return &TunnelController{
		tunnelService: service.GetTunnelService(),
	}
}

func (h *TunnelController) Create(ctx *gin.Context) {
	clientId := getClientID(ctx)
	if clientId == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	var req request.TunnelCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	err := h.tunnelService.Create(ctx, clientId, &req)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

func (h *TunnelController) List(ctx *gin.Context) {
	clientId := getClientID(ctx)
	if clientId == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	tunnels, err := h.tunnelService.ListByClientID(ctx, clientId)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	var list []response.TunnelResponse
	copier.Copy(&list, tunnels)
	response.Success(ctx, list)
}
