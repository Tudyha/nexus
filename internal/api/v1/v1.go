package v1

import (
	"github.com/Tudyha/nexus/internal/middleware"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	// 认证相关
	{
		authController := newAuthController()
		authApi := api.Group("/auth")
		authApi.POST("/send_code", authController.SendCode)
		authApi.POST("/login", middleware.LoginHandler())
	}

	// 用户相关
	{
		userController := newUserController()
		userApi := api.Group("/user").Use(middleware.Auth())
		userApi.GET("/", userController.GetUser)
	}

	// 仪表盘相关
	{
		dashboardController := newDashboardController()
		dashboardApi := api.Group("/dashboard").Use(middleware.Auth())
		dashboardApi.GET("/", dashboardController.GetDashboard)
	}

	// 客户端相关
	{
		clientController := newClientController()
		tunnelController := newTunnelController()
		clientApi := api.Group("/client").Use(middleware.Auth())
		clientApi.GET("/page", clientController.GetPage)
		clientApi.GET("/bind", clientController.GetBind)
		clientApi.GET("/:id", clientController.GetByID)
		clientApi.DELETE("/:id", clientController.Delete)
		clientApi.GET("/:id/pty", clientController.OpenPty)
		clientApi.POST("/v2ray/sub", clientController.GenerateV2raySubscribeLink)
		clientApi.GET("/:id/tunnel", tunnelController.List)
		clientApi.POST("/:id/tunnel", tunnelController.Create)
		clientApi.DELETE("/:id/tunnel/:tunnelId", tunnelController.Delete)
	}
}

func getUserId(ctx *gin.Context) uint64 {
	return ctx.GetUint64(constant.HttpHeaderUserIDKey)
}

func getClientID(ctx *gin.Context) uint64 {
	v, ok := ctx.Params.Get("id")
	if !ok {
		return 0
	}
	return utils.StringToUint64(v)
}

func getAppID(ctx *gin.Context) uint64 {
	return utils.StringToUint64(ctx.Request.Header.Get(constant.HttpHeaderAppIDKey))
}
