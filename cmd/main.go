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
		server.NewMonitor(),
		server.NewTunnelServer(),
	}
	for _, s := range servers {
		if err := s.Start(); err != nil {
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
	log.Info().Msg("servers stopped")
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// init config
	if err := config.Init("./configs/config.yaml"); err != nil {
		log.Fatal().Err(err).Msg("config init failed")
	}
	log.Info().Any("config", config.Get()).Msg("config loaded")

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
