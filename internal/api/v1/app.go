package v1

import (
	"strconv"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/gin-gonic/gin"
)

type AppController struct {
	appService service.AppService
}

func newAppController() *AppController {
	return &AppController{
		appService: service.GetAppService(),
	}
}

func (h *AppController) ListByWorkspace(ctx *gin.Context) {
	workspaceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	list, err := h.appService.ListByWorkspaceID(ctx, workspaceID)
	if err != nil {
		response.FailWithMsg(ctx, err, "查询应用列表失败")
		return
	}
	response.Success(ctx, list)
}

func (h *AppController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的应用ID")
		return
	}
	app, err := h.appService.GetByID(ctx, id)
	if err != nil {
		response.FailWithMsg(ctx, err, "查询应用失败")
		return
	}
	response.Success(ctx, app)
}

func (h *AppController) Create(ctx *gin.Context) {
	workspaceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,max=64"`
		Description string `json:"description" binding:"max=255"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	app, err := h.appService.Create(ctx, workspaceID, req.Name, req.Description)
	if err != nil {
		response.FailWithMsg(ctx, err, "创建应用失败")
		return
	}
	response.Success(ctx, app)
}

func (h *AppController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的应用ID")
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required,max=64"`
		Description string `json:"description" binding:"max=255"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	if err := h.appService.Update(ctx, &model.App{
		BaseModel:   model.BaseModel{ID: id},
		Name:        req.Name,
		Description: req.Description,
	}); err != nil {
		response.FailWithMsg(ctx, err, "更新应用失败")
		return
	}
	response.Success(ctx, nil)
}

func (h *AppController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的应用ID")
		return
	}
	if err := h.appService.Delete(ctx, id); err != nil {
		response.FailWithMsg(ctx, err, "删除应用失败")
		return
	}
	response.Success(ctx, nil)
}

func (h *AppController) UpdateConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的应用ID")
		return
	}
	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	if err := h.appService.UpdateConfig(ctx, id, req.Config); err != nil {
		response.FailWithMsg(ctx, err, "更新配置失败")
		return
	}
	response.Success(ctx, nil)
}
