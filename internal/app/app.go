package app

import (
	"fmt"
	"messanger/config"
	"messanger/internal/repo"
	"messanger/internal/services"
	"messanger/internal/transport/http"
	"messanger/internal/transport/ws"
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

	messageRepo := repo.NewMessageRepo(cfg)
	messageService := services.NewMessageService(messageRepo)
	httpServer := http.NewServer(http.ServerConfig{
		Addr:           cfg.Server.Addr,
		MessageService: messageService,
		Log:            log,
	})

	websocketServer := ws.NewWebSocketServer(messageService, log, cfg.TokenKey)

	go func() {
		if err := httpServer.Run(); err != nil {
			log.Fatal(fmt.Sprintf("error occurred while running HTTP server: %v", err))
		}
	}()
	log.Info(fmt.Sprintf("HTTP server successfully started on %s", cfg.Server.Addr))

	go func() {
		if err := websocketServer.Run(cfg.WSServer.Addr); err != nil {
			log.Error(fmt.Sprintf("error occurred while running WebSocket server: %v", err))
		}
	}()
	log.Info(fmt.Sprintf("WebSocket server successfully started on %s", cfg.WSServer.Addr))

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
