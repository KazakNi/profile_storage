package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"users/config"
	delivery "users/internal/user/infrastructure/delivery/http"
	"users/internal/user/infrastructure/repository"
	storage "users/pkg/db"
	slogger "users/pkg/logger"
)

func Run() {

	slogger.Logger = slogger.GetLogger()

	userdb, authdb := storage.InMemoryStorage{Storage: make(map[string][]byte)}, storage.InMemoryStorage{Storage: make(map[string][]byte)}

	UserRepo := repository.NewBannerRepository(&userdb, &authdb)
	UserRepo.CreateAdmin()
	UserHandler := delivery.NewUserHandler(UserRepo)

	mux := http.NewServeMux()

	mux.Handle("/user/", UserHandler)

	// Swagger specification

	mux.HandleFunc("GET /redoc", delivery.ReDoc)
	mux.Handle("/swagger.yaml", http.FileServer(http.Dir(config.Cfg.Swagger.StaticPath)))

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.Cfg.Server.Host, config.Cfg.Server.Port), mux); err != nil {
			slogger.Logger.Error("error, server is crashed: ", "err", err)
		}
	}()

	slogger.Logger.Info("Listening to", "HOST", config.Cfg.Server.Host, "PORT", config.Cfg.Server.Port)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	s := <-sigChan
	slogger.Logger.Info("Shutdown server", "signal", s)
}
