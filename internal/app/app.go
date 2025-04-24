package app

import (
	"fmt"
	"messanger/config"
	"messanger/internal/repo"
	"messanger/internal/services"
	"messanger/internal/transport/http"
	"messanger/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	log := logger.GetLogger()

	log.Info("Staring messanger-app...")

	db := repo.ConnectDB(cfg)

	messageRepo := repo.NewMessageRepo(db)
	messageService := services.NewMessageService(messageRepo)
	httpServer := http.NewServer(http.ServerConfig{
		Addr:           cfg.Server.Addr,
		MessageService: messageService,
		Log:            log,
	})

	go func() {
		if err := httpServer.Run(); err != nil {
			log.Fatal(fmt.Sprintf("error occurred while running HTTP server: %v", err))
		}
	}()
	log.Info(fmt.Sprintf("HTTP server successfully started on %s", cfg.Server.Addr))

	log.Info("Application started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	<-quit

	log.Info("shutdown HTTP server...")
	if err := httpServer.Shutdown(); err != nil {
		log.Error(fmt.Sprintf("failed to shutdown HTTP server: %v", err))
	} else {
		log.Info("HTTP server successfully shutdown")
	}
}
