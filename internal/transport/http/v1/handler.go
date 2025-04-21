package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"messanger/internal/services"
)

type Handler struct {
	messageService services.MessageService
	log            *logrus.Logger
}

type HandlerConfig struct {
	MessageService services.MessageService
	Log            *logrus.Logger
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		messageService: cfg.MessageService,
		log:            cfg.Log,
	}
}

func (h *Handler) Init(router fiber.Router) {
	h.initMessageRoutes(router.Group("/v1"))
}
