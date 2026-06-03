package api

import (
	"net/http"
	"path/filepath"
	"strings"

	v1 "github.com/Tudyha/nexus/internal/api/v1"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "ok",
			"data": nil,
		})
	})

	// Prometheus 指标
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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

		// 防止路径遍历攻击
		if filename == "" || strings.Contains(filename, "..") || filename[0] == '/' || filename[0] == '\\' {
			ctx.JSON(http.StatusOK, gin.H{
				"code": 400,
				"msg":  "invalid filename",
				"data": nil,
			})
			return
		}

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
