package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	_  "github.com/satrunjis/user-service/docs"
	"github.com/satrunjis/user-service/internal/config"
	"github.com/satrunjis/user-service/internal/external/maptile"
	"github.com/satrunjis/user-service/internal/logger"
	"github.com/satrunjis/user-service/internal/mapcache"
	"github.com/satrunjis/user-service/internal/repository/elastic"
	"github.com/satrunjis/user-service/internal/server"
)

// @title User Service API
// @version 1.0
// @description Сервис управления пользователями с интеграцией Elasticsearch, сервиса получения map tile и их кеширования с помощью Redis

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	//docs.SwaggerInfo.Host = cfg.HTTPServerConfig.Address

	logger := logger.New(cfg.Env, nil)

	mapService := maptile.Init(logger)
	cacheService, err := mapcache.Init(ctx, &cfg.CacheConfig.URL, logger, cfg.CacheConfig.TTL)
	if err != nil {
		logger.Error("Failed to initialize cache service", "err", err)
		return
	}

	esClient, err := elastic.Init(ctx, cfg.ElasticConfig.URL, logger)
	if err != nil {
		logger.Error("Failed to initialize Elasticsearch client", "err", err)
		return
	}

	serv := server.NewServer(&cfg.HTTPServerConfig, logger, esClient, cacheService, mapService)

	go func() {
		if err := serv.Run(); err != nil {
			logger.Error("HTTP server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Received shutdown signal, initiating graceful shutdown...")
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	errorsOccurred := false

	if err := serv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "err", err)
	}

	if err := esClient.Close(); err != nil {
		logger.Error("Failed to close Elasticsearch client", "err", err)
	}

	if err := cacheService.Close(); err != nil {
		logger.Error("Failed to close Redis cache service", "err", err)
	}
	if !errorsOccurred {
		logger.Info("Server exited gracefully")
	} else {
		logger.Error("Server exited with errors during shutdown")
	}
}
