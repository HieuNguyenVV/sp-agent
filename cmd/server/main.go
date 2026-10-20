package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HieuNguyenVV/sp-agent/internal/pkg/config"
	"github.com/HieuNguyenVV/sp-agent/internal/server"
)

// @title           SP Agent API
// @version         1.0
// @description     SP Agent REST API
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey AccessToken
// @in              header
// @name            Authorization
// @securityDefinitions.apikey RefreshToken
// @in              header
// @name            RefreshToken
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	clientContainer := server.NewClientContainer(cfg)
	repositoryContainer := server.NewRepositoryContainer(clientContainer)
	serviceContainer := server.NewServiceContainer(clientContainer, repositoryContainer)
	controllerContainer := server.NewControllerContainer(clientContainer, serviceContainer)
	srv := server.NewServer(clientContainer, repositoryContainer, serviceContainer, controllerContainer)
	srv.Run(ctx)
}
