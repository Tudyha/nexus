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
	v2rayController := newV2RayController()
	appController := newAppController()

	// 认证相关
	{
		authApi := api.Group("/auth")
		authApi.POST("/send_code", authController.SendCode)
		authApi.POST("/login", middleware.LoginHandler())
		authApi.POST("/logout", middleware.Auth(), authController.Logout)
		authApi.POST("/refresh", authController.RefreshToken)
	}

	// 用户相关
	{
		userApi := api.Group("/user").Use(middleware.Auth())
		userApi.GET("/", userController.GetUser)
		userApi.GET("/list", userController.List)
		userApi.PUT("/password", authController.SetPassword)
	}

	// 仪表盘相关
	{
		dashboardApi := api.Group("/dashboard").Use(middleware.Auth())
		dashboardApi.GET("/", dashboardController.GetDashboard)
	}

	// 客户端相关
	{

		clientApi := api.Group("/client").Use(middleware.Auth())
		// 客户端基础操作
		clientApi.GET("/page", clientController.GetPage)
		clientApi.GET("/:id", clientController.GetByID)
		clientApi.DELETE("/:id", clientController.Delete)
		clientApi.GET("/online", clientController.ListOnline)
		clientApi.GET("/config", clientController.GetConfig)
		clientApi.PUT("/config", clientController.UpdateConfig)

		clientApi.GET("/bind", clientController.GetBind)
		clientApi.GET("/:id/terminal", clientController.Terminal)
		clientApi.POST("/:id/upgrade", clientController.Upgrade)

		// 客户端进程管理
		clientApi.GET("/:id/processes", clientController.ProcessList)
		clientApi.DELETE("/:id/processes/:pid", clientController.ProcessKill)

		// 客户端文件管理
		clientApi.GET("/:id/files", clientController.FileList)
		clientApi.GET("/:id/files/download", clientController.FileDownload)
		clientApi.POST("/:id/files/upload", clientController.FileUpload)
		clientApi.DELETE("/:id/files/delete", clientController.FileDelete)
		clientApi.POST("/:id/files/mkdir", clientController.FileMkdir)
		clientApi.POST("/:id/files/rename", clientController.FileRename)

		// 客户端网络管理
		clientApi.GET("/:id/network", clientController.NetworkList)

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
		taskApi.POST("/create", taskController.Create)
		taskApi.GET("/:taskId/:clientId", taskController.GetExecution)
	}

	// 隧道管理
	{
		tunnelApi := api.Group("/tunnel").Use(middleware.Auth())
		tunnelApi.GET("/list", tunnelController.ListAll)
		tunnelApi.POST("/", tunnelController.Create)
	}

	// 工作空间管理
	{
		workspaceApi := api.Group("/workspace").Use(middleware.Auth())
		workspaceController := newWorkspaceController()
		workspaceApi.GET("", workspaceController.List)
		workspaceApi.GET("/:id", workspaceController.GetByID)
		workspaceApi.POST("", workspaceController.Create)
		workspaceApi.PUT("/:id", workspaceController.Update)
		workspaceApi.DELETE("/:id", workspaceController.Delete)

		// 工作空间用户管理
		workspaceApi.GET("/:id/users", workspaceController.ListUsers)
		workspaceApi.POST("/:id/users", workspaceController.AddUser)
		workspaceApi.DELETE("/:id/users/:userId", workspaceController.RemoveUser)
		workspaceApi.PUT("/:id/users/:userId/role", workspaceController.UpdateUserRole)

		// 应用管理（嵌套在工作空间下）
		workspaceApi.GET("/:id/apps", appController.ListByWorkspace)
		workspaceApi.POST("/:id/apps", appController.Create)
	}

	// 应用管理（独立路径）
	{
		appApi := api.Group("/app").Use(middleware.Auth())
		appApi.GET("/:id", appController.GetByID)
		appApi.PUT("/:id", appController.Update)
		appApi.DELETE("/:id", appController.Delete)
		appApi.PUT("/:id/config", appController.UpdateConfig)
	}

	// v2ray
	{
		v2rayApi := api.Group("/v2ray")
		v2rayApi.GET("/:appId/sub", v2rayController.V2raySubscribe)
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
