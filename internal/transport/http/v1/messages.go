package v1

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"messanger/internal/services"
	"strconv"
)

func (h *Handler) initMessageRoutes(router fiber.Router) {
	messages := router.Group("/messages")
	{
		messages.Get("/:id", h.getMessagesByID)
		messages.Post("/", h.saveMessage)
	}
}

func (h *Handler) getMessagesByID(c *fiber.Ctx) error {
	receiverID := c.Params("id")
	if receiverID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "receiverID is required")
	}

	receiverIDInt, err := strconv.Atoi(receiverID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid receiverID: %v", err))
	}

	messages, err := h.messageService.GetHistory(context.Background(), services.GetHistoryParams{
		SenderID:   0,
		ReceiverID: uint32(receiverIDInt),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("h.messageService.GetHistory: %v", err))
	}

	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "no messages found")
	}

	return c.JSON(messages)
}

type CreateMessageRequest struct {
	SenderID   uint32 `json:"sender_id"`
	ReceiverID uint32 `json:"receiver_id"`
	Content    string `json:"content"`
}

func (h *Handler) saveMessage(c *fiber.Ctx) error {
	var req CreateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request body")
	}

	err := h.messageService.SaveMessage(context.Background(), services.SaveMessageParams{
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("h.messageService.SaveMessage: %v", err))
	}

	return c.SendStatus(fiber.StatusCreated)
}
