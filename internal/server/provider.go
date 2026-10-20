package server

import (
	"github.com/HieuNguyenVV/sp-agent/internal/log"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/cache"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/config"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/database"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/jwt"
	"github.com/HieuNguyenVV/sp-agent/internal/server/middleware"
	"github.com/HieuNguyenVV/sp-agent/internal/server/validator"
	"github.com/gin-gonic/gin"
)

type ClientContainer struct {
	config     *config.Config
	logger     log.Logger
	httpClient *gin.Engine
	postgres   *database.Postgres
	redis      *cache.Redis
}

type ControllerContainer struct{}

type ServiceContainer struct {
	authMiddleware         gin.HandlerFunc
	refreshTokenMiddleware gin.HandlerFunc
	jwtService             jwt.JWTService
}

type RepositoryContainer struct{}

func NewClientContainer(cfg *config.Config) *ClientContainer {
	logger := log.New(&log.Config{
		LogLevel: cfg.Log.LogLevel,
	})

	postgres, err := database.NewPostgres(cfg.Postgres)
	if err != nil {
		logger.Error("failed to create postgres client", "error", err)
		panic(err)
	}

	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		logger.Error("failed to create redis client", "error", err)
		panic(err)
	}

	httpClient := NewHTTPServer(cfg, logger)

	return &ClientContainer{
		config:     cfg,
		logger:     logger,
		httpClient: httpClient,
		postgres:   postgres,
		redis:      redisClient,
	}
}

func NewHTTPServer(cfg *config.Config, logger log.Logger) *gin.Engine {
	validator.Register()

	var engine *gin.Engine
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
		engine = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	}

	engine.Use(
		middleware.Cors(),
		middleware.Tx(logger),
		gin.Recovery(),
	)
	return engine
}

func NewServiceContainer(clientContainer *ClientContainer, _ *RepositoryContainer) *ServiceContainer {
	jwtService := jwt.NewJWTService(clientContainer.config)
	return &ServiceContainer{
		authMiddleware:         middleware.NewAuthMiddleware(clientContainer.logger, jwtService),
		refreshTokenMiddleware: middleware.NewRefreshTokenMiddleware(clientContainer.logger, jwtService),
		jwtService:             jwtService,
	}
}

func NewRepositoryContainer(_ *ClientContainer) *RepositoryContainer {
	return &RepositoryContainer{}
}

func NewControllerContainer(_ *ClientContainer, _ *ServiceContainer) *ControllerContainer {
	return &ControllerContainer{}
}
