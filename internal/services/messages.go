package services

import (
	"context"
	"fmt"
	"golang.org/x/net/websocket"
	"messanger/internal/models"
	"messanger/internal/repo"
)

type MessageService interface {
	SaveMessage(ctx context.Context, params SaveMessageParams) error
	GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error)
}

type messageService struct {
	repo repo.MessageRepo
	ws   *websocket.Conn
}

func NewMessageService(repo repo.MessageRepo) MessageService {
	return &messageService{repo: repo}
}

type SaveMessageParams struct {
	SenderID   int64
	ReceiverID int64
	Content    string
}

func (s *messageService) SaveMessage(ctx context.Context, params SaveMessageParams) error {
	if err := s.repo.SaveMessage(ctx, repo.SaveMessageParams{
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
		Content:    params.Content,
	}); err != nil {
		return fmt.Errorf("s.repo.SaveMessage: %w", err)
	}
	return nil
}

type GetHistoryParams struct {
	SenderID   int64
	ReceiverID int64
}

func (s *messageService) GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error) {
	messages, err := s.repo.GetHistory(ctx, repo.GetHistoryParams{
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
	})
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetHistory: %w", err)
	}

	return messages, nil
}
