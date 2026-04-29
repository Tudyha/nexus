package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
	registerMiddlewares(router)

	// 注册路由
	api.RegisterRoutes(router)

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
			return
		}
	}()

	log.Info().Str("addr", h.s.Addr).Msg("http server started")
	return nil
}

func (h *httpServer) Stop() error {
	if h.s == nil {
		return nil
	}
	return h.s.Shutdown(context.Background())
}

func (h *httpServer) String() string {
	return "HTTP Server"
}

// registerMiddlewares 注册中间件
func registerMiddlewares(router *gin.Engine) {
	// 基础中间件
	router.Use(gin.Recovery())

	// 自定义中间件
	middleware.Init()
}
