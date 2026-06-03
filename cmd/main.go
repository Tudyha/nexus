package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/dao"
	"github.com/Tudyha/nexus/internal/database"
	"github.com/Tudyha/nexus/internal/mq"
	"github.com/Tudyha/nexus/internal/server"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	servers := []server.Server{
		server.NewHTTPServer(),
		server.NewTCPServer(),
		server.NewV2rayServer(),
		server.NewTunnelServer(),
	}
	for i, s := range servers {
		if err := s.Start(); err != nil {
			// 启动失败时关闭已启动的服务器
			for j := range i {
				if e := servers[j].Stop(); e != nil {
					log.Error().Err(e).Msg("server stop failed")
				}
			}
			log.Fatal().Err(err).Msg("server start failed")
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down servers...")
	for _, s := range servers {
		if err := s.Stop(); err != nil {
			log.Error().Err(err).Msg("server stop failed")
		}
	}
	if err := mq.Close(); err != nil {
		log.Error().Err(err).Msg("mq close failed")
	}
	log.Info().Msg("servers stopped")
}

func init() {
	// init config
	if err := config.Init("./configs/config.yaml"); err != nil {
		log.Fatal().Err(err).Msg("config init failed")
	}
	cfg := config.Get()

	// 根据环境配置日志格式
	if cfg.Server.Env == "prod" {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	// 配置日志级别
	switch cfg.Server.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Debug().Msg("log level set to " + cfg.Server.LogLevel)

	log.Info().Any("config", cfg).Msg("config loaded")

	// init database
	if err := database.Init(); err != nil {
		log.Fatal().Err(err).Msg("database init failed")
	}

	// init dao
	if err := dao.Init(); err != nil {
		log.Fatal().Err(err).Msg("dao init failed")
	}

	// init mq
	if err := mq.Init(); err != nil {
		log.Fatal().Err(err).Msg("mq init failed")
	}

	// init service
	if err := service.Init(); err != nil {
		log.Fatal().Err(err).Msg("service init failed")
	}

	// init session manager
	if err := session.Init(); err != nil {
		log.Fatal().Err(err).Msg("session manager init failed")
	}
}
