package v1

import (
	"net/http"
	"strconv"

	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/gin-gonic/gin"
)

type WorkspaceController struct {
	workspaceService service.WorkspaceService
	userService      service.UserService
}

func newWorkspaceController() *WorkspaceController {
	return &WorkspaceController{
		workspaceService: service.GetWorkspaceService(),
		userService:      service.GetUserService(),
	}
}

func (h *WorkspaceController) List(ctx *gin.Context) {
	list, err := h.workspaceService.List(ctx)
	if err != nil {
		response.FailWithMsg(ctx, err, "查询工作空间列表失败")
		return
	}
	response.Success(ctx, list)
}

func (h *WorkspaceController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	space, err := h.workspaceService.GetByID(ctx, id)
	if err != nil {
		response.FailWithMsg(ctx, err, "查询工作空间失败")
		return
	}
	response.Success(ctx, space)
}

func (h *WorkspaceController) Create(ctx *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=64"`
		Description string `json:"description" binding:"max=255"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}

	space, err := h.workspaceService.Create(ctx, req.Name, req.Description)
	if err != nil {
		response.FailWithMsg(ctx, err, "创建工作空间失败")
		return
	}

	// 将创建者自动加入为该工作空间的管理员
	userID := getUserId(ctx)
	if userID > 0 {
		_ = h.workspaceService.AddUser(ctx, space.ID, userID, 1)
	}

	response.Success(ctx, space)
}

func (h *WorkspaceController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
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
	if err := h.workspaceService.Update(ctx, &model.Workspace{
		BaseModel:   model.BaseModel{ID: id},
		Name:        req.Name,
		Description: req.Description,
	}); err != nil {
		response.FailWithMsg(ctx, err, "更新工作空间失败")
		return
	}
	response.Success(ctx, nil)
}

func (h *WorkspaceController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	if err := h.workspaceService.Delete(ctx, id); err != nil {
		response.FailWithMsg(ctx, err, "删除工作空间失败")
		return
	}
	response.Success(ctx, nil)
}

func (h *WorkspaceController) ListUsers(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	users, err := h.workspaceService.ListUsers(ctx, id)
	if err != nil {
		response.FailWithMsg(ctx, err, "查询用户列表失败")
		return
	}
	response.Success(ctx, users)
}

func (h *WorkspaceController) AddUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Role   int    `json:"role" binding:"required,oneof=1 2"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	if err := h.workspaceService.AddUser(ctx, id, req.UserID, req.Role); err != nil {
		response.FailWithMsg(ctx, err, "添加用户失败")
		return
	}
	ctx.JSON(http.StatusOK, response.Response{Code: 0, Msg: "ok"})
}

func (h *WorkspaceController) RemoveUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	userID, err := strconv.ParseUint(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的用户ID")
		return
	}
	if err := h.workspaceService.RemoveUser(ctx, id, userID); err != nil {
		response.FailWithMsg(ctx, err, "移除用户失败")
		return
	}
	ctx.JSON(http.StatusOK, response.Response{Code: 0, Msg: "ok"})
}


func (h *WorkspaceController) UpdateUserRole(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的工作空间ID")
		return
	}
	userID, err := strconv.ParseUint(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.FailWithMsg(ctx, nil, "无效的用户ID")
		return
	}
	var req struct {
		Role int `json:"role" binding:"required,oneof=1 2"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, nil, "参数错误: "+err.Error())
		return
	}
	if err := h.workspaceService.UpdateUserRole(ctx, id, userID, req.Role); err != nil {
		response.FailWithMsg(ctx, err, "更新角色失败")
		return
	}
	ctx.JSON(http.StatusOK, response.Response{Code: 0, Msg: "ok"})
}
