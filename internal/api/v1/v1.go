package v1

import (
	"github.com/Tudyha/nexus/internal/middleware"
	constant "github.com/Tudyha/nexus/pkg/const"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	authController := newAuthController()
	userController := newUserController()
	dashboardController := newDashboardController()
	clientController := newClientController()
	tunnelController := newTunnelController()
	versionController := newVersionController()
	taskController := newTaskController()

	// 认证相关
	{
		authApi := api.Group("/auth")
		authApi.POST("/send_code", authController.SendCode)
		authApi.POST("/login", middleware.LoginHandler())
	}

	// 用户相关
	{
		userApi := api.Group("/user").Use(middleware.Auth())
		userApi.GET("/", userController.GetUser)
	}

	// 仪表盘相关
	{
		dashboardApi := api.Group("/dashboard").Use(middleware.Auth())
		dashboardApi.GET("/", dashboardController.GetDashboard)
	}

	// 客户端相关
	{

		clientApi := api.Group("/client").Use(middleware.Auth())
		clientApi.GET("/page", clientController.GetPage)
		clientApi.GET("/bind", clientController.GetBind)
		clientApi.GET("/:id", clientController.GetByID)
		clientApi.DELETE("/:id", clientController.Delete)
		clientApi.GET("/:id/terminal", clientController.Terminal)
		clientApi.POST("/v2ray/sub", clientController.GenerateV2raySubscribeLink)
		clientApi.GET("/:id/tunnel", tunnelController.List)
		clientApi.POST("/:id/tunnel", tunnelController.Create)
		clientApi.DELETE("/:id/tunnel/:tunnelId", tunnelController.Delete)
		clientApi.POST("/:id/upgrade", versionController.Upgrade)
	}

	// 版本管理
	{
		versionApi := api.Group("/version")
		versionApi.POST("", middleware.Auth(), versionController.Upload)
		versionApi.GET("/page", middleware.Auth(), versionController.Page)
		versionApi.DELETE("/:id", middleware.Auth(), versionController.Delete)
		versionApi.GET("/latest", versionController.Latest)
	}

	// 任务管理
	{
		taskApi := api.Group("/task").Use(middleware.Auth())
		taskApi.GET("/:taskId/:clientId", taskController.GetExecution)
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
