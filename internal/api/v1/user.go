package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
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
