package app

import (
	"github.com/shamank/edutour-backend/apigateway/internal/config"
	"github.com/shamank/edutour-backend/apigateway/internal/delivery/http"
	"github.com/shamank/edutour-backend/apigateway/internal/server"
	"github.com/shamank/edutour-backend/apigateway/pkg/logger/sl"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func Run(configPath string) {
	// TODO: init config
	cfg := config.InitConfig(configPath)

	// TODO: init logger
	logger := sl.SetupLogger(cfg.Env)

	authServiceAddr := cfg.AuthService.Http.Schema + "://" + cfg.AuthService.Http.Host + ":" + strconv.Itoa(cfg.AuthService.Http.Port)
	dataServiceAddr := cfg.DataService.Http.Schema + "://" + cfg.DataService.Http.Host + ":" + strconv.Itoa(cfg.DataService.Http.Port)

	// TODO: init routes
	handlers := http.NewHandler(logger, http.Services{
		//AuthServiceAddr:    "http://" + cfg.AuthService.Host + ":" + strconv.Itoa(cfg.AuthService.Port),
		//BackendServiceAddr: "http://" + cfg.BackendService.Host + ":" + strconv.Itoa(cfg.BackendService.Port),
		AuthServiceAddr: authServiceAddr,
		DataServiceAddr: dataServiceAddr,
	})

	// TODO: run http server
	srv := server.NewServer(cfg.HTTPServer, handlers.InitHandle())

	go func() {
		logger.Info("[env: " + cfg.Env + "] HTTP-server start up!")
		if err := srv.Start(); err != nil {
			logger.Error("error occurred when starting the HTTP-server", sl.Err(err))
			return
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := srv.Stop(); err != nil {
		logger.Error("error occurred when stoppiing HTTP-server", sl.Err(err))
		return
	}
	logger.Info("HTTP-server has shut down")
}
