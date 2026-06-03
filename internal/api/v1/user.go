package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func newUserController() *UserController {
	return &UserController{
		userService: service.GetUserService(),
	}
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.UserResponse "成功响应"
// @Router /user [get]
func (h *UserController) GetUser(ctx *gin.Context) {
	userId := getUserId(ctx)
	user, err := h.userService.Detail(ctx, userId)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, user)
}

func (h *UserController) List(ctx *gin.Context) {
	page := utils.StringToUint64(ctx.DefaultQuery("page", "1"))
	pageSize := utils.StringToUint64(ctx.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	result, err := h.userService.List(ctx, int(page), int(pageSize))
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, result)
}
