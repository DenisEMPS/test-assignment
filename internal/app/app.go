package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DenisEMPS/test-assignment/internal/config"
	"github.com/DenisEMPS/test-assignment/internal/delivery/http"
	"github.com/DenisEMPS/test-assignment/internal/repository/postgres"
	"github.com/DenisEMPS/test-assignment/internal/server"
	"github.com/DenisEMPS/test-assignment/internal/service/auth"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("application starting", slog.String("env", cfg.Env))

	db, err := postgres.NewDB(postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.Username,
		Password: cfg.Postgres.Password,
		DBName:   cfg.Postgres.DBname,
		SSLMode:  cfg.Postgres.SSLmode,
	})
	if err != nil {
		log.Error("failed to init postgres connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	repo := postgres.NewAuth(db)
	service := auth.New(repo, log, &cfg.JWT)
	handler := http.NewHandler(service)

	serv := new(server.Server)

	go func() {
		if err := serv.Run(handler.InitRoutes(), cfg.Server.Port); err != nil {
			log.Error("failed to run server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	log.Info("application started", slog.String("port", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := serv.Shutdown(context.Background()); err != nil {
		log.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	if err := db.Close(); err != nil {
		log.Error("failed to close postgres connection", slog.String("error", err.Error()))
	}

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
