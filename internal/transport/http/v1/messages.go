package v1

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"messanger/internal/services"
	"strconv"
)

// HTTPError представляет ошибку HTTP-ответа
type HTTPError struct {
	Message string `json:"message"`
}

// MessageResponse DTO для ответа в API
type MessageResponse struct {
	ID         int64  `json:"id"`
	SenderID   int64  `json:"sender_id"`
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

func (h *Handler) initMessageRoutes(router fiber.Router) {
	messages := router.Group("/messages")
	{
		messages.Get("/:id", h.GetMessagesByID)
		messages.Post("/", h.CreateMessage)
	}
}

type CreateMessageRequest struct {
	SenderID   int64  `json:"sender_id"`
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
}

// CreateMessage создаёт новое сообщение
// @Summary Создать сообщение
// @Tags messages
// @Description Сохраняет новое сообщение между двумя пользователями
// @Accept json
// @Produce json
// @Param message body CreateMessageRequest true "Данные сообщения"
// @Success 201 {string} string "Created"
// @Failure 400 {object} HTTPError "Ошибка при парсинге запроса"
// @Failure 500 {object} HTTPError "Ошибка при сохранении сообщения"
// @Router /messages [post]
func (h *Handler) CreateMessage(c *fiber.Ctx) error {
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

// GetMessagesByID возвращает историю сообщений между текущим пользователем и получателем
// @Summary Получить сообщения по ID получателя
// @Tags messages
// @Description Возвращает историю сообщений между текущим пользователем (в будущем — по токену) и указанным получателем
// @Param id path int true "ID получателя"
// @Produce json
// @Success 200 {array} MessageResponse
// @Failure 400 {object} HTTPError "Неверный ID"
// @Failure 404 {object} HTTPError "Сообщения не найдены"
// @Failure 500 {object} HTTPError "Внутренняя ошибка сервера"
// @Router /messages/{id} [get]
func (h *Handler) GetMessagesByID(c *fiber.Ctx) error {
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
		ReceiverID: int64(receiverIDInt),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("h.messageService.GetHistory: %v", err))
	}

	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "no messages found")
	}

	return c.JSON(messages)
}
