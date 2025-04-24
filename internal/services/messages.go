package services

import (
	"context"
	"fmt"
	"messanger/internal/models"
	"messanger/internal/repo/pg"
)

type MessageService interface {
	SaveMessage(ctx context.Context, params SaveMessageParams) error
	GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error)
}

type messageService struct {
	pg pg.MessageRepo
}

func NewMessageService(pg pg.MessageRepo) MessageService {
	return &messageService{pg: pg}
}

type SaveMessageParams struct {
	SenderID   uint32
	ReceiverID uint32
	Content    string
}

func (s *messageService) SaveMessage(ctx context.Context, params SaveMessageParams) error {
	if err := s.pg.SaveMessage(ctx, pg.SaveMessageParams{
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
		Content:    params.Content,
	}); err != nil {
		return fmt.Errorf("s.pg.SaveMessage: %w", err)
	}
	return nil
}

type GetHistoryParams struct {
	SenderID   uint32
	ReceiverID uint32
}

func (s *messageService) GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error) {
	messages, err := s.pg.GetHistory(ctx, pg.GetHistoryParams{
		SenderID:   params.SenderID,
		ReceiverID: params.ReceiverID,
	})
	if err != nil {
		return nil, fmt.Errorf("s.pg.GetHistory: %w", err)
	}

	return messages, nil
}
