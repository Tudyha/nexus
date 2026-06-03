package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/Tudyha/nexus/internal/api"
	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type httpServer struct {
	s *http.Server
}

func NewHTTPServer() Server {
	cfg := config.Get()

	router := gin.New()

	// 注册中间件
	registerMiddlewares(router, cfg.Server.JWTSecret)

	// 注册路由
	api.RegisterRoutes(router)

	pprofRouter := router.Group("/debug")
	pprofRouter.GET("/pprof/*profile", gin.WrapH(http.DefaultServeMux))

	addr := net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", cfg.Server.HTTP.Port))
	s := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.HTTP.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &httpServer{s: s}
}

func (h *httpServer) Start() error {
	go func() {
		if err := h.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Str("addr", h.s.Addr).Msg("http server listen error")
		}
	}()

	log.Info().Str("addr", h.s.Addr).Msg("http server started")
	return nil
}

func (h *httpServer) Stop() error {
	if h.s == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return h.s.Shutdown(ctx)
}

func (h *httpServer) String() string {
	return "HTTP Server"
}

// registerMiddlewares 注册中间件
func registerMiddlewares(router *gin.Engine, jwtSecret string) {
	// 基础中间件
	router.Use(gin.Recovery())

	// 自定义中间件
	middleware.Init(jwtSecret)

	// 速率限制
	cfg := config.Get()
	if cfg.Server.RateLimit.Enabled {
		middleware.InitRateLimiter(cfg.Server.RateLimit.Rate, cfg.Server.RateLimit.Burst)
		router.Use(middleware.RateLimit())
	}
}
