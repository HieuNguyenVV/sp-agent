package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	clientContainer     *ClientContainer
	repositoryContainer *RepositoryContainer
	serviceContainer    *ServiceContainer
	controllerContainer *ControllerContainer
}

func NewServer(
	clientContainer *ClientContainer,
	repositoryContainer *RepositoryContainer,
	serviceContainer *ServiceContainer,
	controllerContainer *ControllerContainer,
) *Server {
	engine := clientContainer.httpClient
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if err := clientContainer.postgres.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	return &Server{
		clientContainer:     clientContainer,
		repositoryContainer: repositoryContainer,
		serviceContainer:    serviceContainer,
		controllerContainer: controllerContainer,
	}
}

func (s *Server) Run(ctx context.Context) {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", s.clientContainer.config.App.Port),
		Handler:           s.clientContainer.httpClient,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		s.clientContainer.logger.Infof("starting server on port %s", s.clientContainer.config.App.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.clientContainer.logger.WithError(err).Error("failed to start server")
		}
	}()

	<-ctx.Done()

	s.clientContainer.logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		s.clientContainer.logger.WithError(err).Error("failed to shutdown server")
	}
	s.clientContainer.postgres.Shutdown()
	if s.clientContainer.redis != nil {
		if err := s.clientContainer.redis.Shutdown(); err != nil {
			s.clientContainer.logger.WithError(err).Error("failed to shutdown redis")
		}
	}
	s.clientContainer.logger.Info("server shutdown complete")
}
