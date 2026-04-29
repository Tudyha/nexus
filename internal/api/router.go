package api

import (
	"net/http"
	"path/filepath"

	v1 "github.com/Tudyha/nexus/internal/api/v1"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(router *gin.Engine) {
	registerStaticRoutes(router)
	v1.RegisterRoutes(router.Group("/api/v1"))
}

// 注册静态路由
func registerStaticRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "ok",
			"data": nil,
		})
	})

	// 静态文件
	router.GET("/:name", func(ctx *gin.Context) {
		filename := ctx.Params.ByName("name")
		if filename == "" {
			ctx.JSON(http.StatusOK, gin.H{
				"code": 400,
				"msg":  "filename is empty",
				"data": nil,
			})
			return
		}

		filename = filepath.Clean(filename)

		if utils.FileExists("./tmp/" + filename) {
			ctx.File("./tmp/" + filename)
			return
		}

		ctx.File("./build/" + filename)
	})

	//前端静态文件
	router.Static("/assets", "web/dist/assets")
	router.GET("/", func(c *gin.Context) {
		c.File("web/dist/index.html")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("web/dist/favicon.ico")
	})
}
