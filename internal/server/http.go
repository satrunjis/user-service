package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satrunjis/user-service/internal/config"
	"github.com/satrunjis/user-service/internal/domain"
	"github.com/satrunjis/user-service/internal/handler"
	"github.com/satrunjis/user-service/internal/middleware"
	"github.com/satrunjis/user-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	logger     *slog.Logger
	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(
	cfg *config.HTTPServerConfig,
	logger *slog.Logger,
	userRepo domain.UserRepository,
	cacheService service.CacheService,
	mapService service.MapService,
) *Server {
	userService := service.NewUserService(userRepo, cacheService, mapService)
	userHandler := handler.NewUserHandler(userService, logger)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.Cors())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
		)
	})

	setupRoutes(router, userHandler)

	httpServer := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		logger:     logger,
		httpServer: httpServer,
		router:     router,
	}
}
func setupRoutes(router *gin.Engine, userHandler *handler.UserHandler) {
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	group := router.Group("/api/v1")
	{
		users := group.Group("/users")
		users.GET("", userHandler.GetUsers)
		users.POST("", userHandler.CreateUser)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.PATCH("/:id", userHandler.UpdateUserPartial)
		users.DELETE("/:id", userHandler.DeleteUser)
		users.GET("/:id/map", userHandler.GetUserMap)
	}
	router.GET("/health", healthCheck)

}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
func (s *Server) Run() error {

	s.logger.Info("Starting server", "addr", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Failed to start server", "err", err)
		return err
	}

	return nil
}
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
