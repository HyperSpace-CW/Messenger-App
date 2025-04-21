package http

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"
	"messanger/internal/services"
	v1 "messanger/internal/transport/http/v1"
)

type Server struct {
	addr string

	messageService services.MessageService

	log *logrus.Logger
	app *fiber.App
}

type ServerConfig struct {
	Addr string

	MessageService services.MessageService

	Log *logrus.Logger
}

func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		addr:           cfg.Addr,
		messageService: cfg.MessageService,
		log:            cfg.Log,
	}

	server.app = fiber.New(fiber.Config{})

	server.init()

	return server
}

func (s *Server) Run() error {
	if err := s.app.Listen(s.addr); err != nil {
		return fmt.Errorf("listening HTTP server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	if err := s.app.Shutdown(); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}

func (s *Server) init() {
	s.app.Use(cors.New())
	s.app.Use(requestid.New())

	s.setHandlers()
}

func (s *Server) setHandlers() {
	handlerV1 := v1.NewHandler(v1.HandlerConfig{
		MessageService: s.messageService,
		Log:            s.log,
	})
	{
		handlerV1.Init(s.app)
	}
}
