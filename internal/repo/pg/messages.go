package pg

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"messanger/internal/models"
	"sync"
	"time"
)

type MessageRepo interface {
	SaveMessage(ctx context.Context, params SaveMessageParams) error
	GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error)
}

type messageRepo struct {
	db *sqlx.DB
	mu *sync.RWMutex
}

func NewMessageRepo(db *sqlx.DB) MessageRepo {
	return &messageRepo{
		db: db,
		mu: new(sync.RWMutex),
	}
}

type message struct {
	ID         string     `db:"id"`
	SenderID   string     `db:"sender_id"`
	ReceiverID string     `db:"receiver_id"`
	Content    string     `db:"content"`
	SentAt     *time.Time `db:"sent_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

type SaveMessageParams struct {
	SenderID   uint32
	ReceiverID uint32
	Content    string
}

const saveMessageQuery = `
INSERT INTO messages (sender_id, receiver_id, content, sent_at, created_at)
VALUES ($1, $2, $3, $4, $5)
`

func (m messageRepo) SaveMessage(ctx context.Context, params SaveMessageParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.db.ExecContext(
		ctx,
		saveMessageQuery,
		params.SenderID,
		params.ReceiverID,
		params.Content,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

type GetHistoryParams struct {
	SenderID   uint32
	ReceiverID uint32
}

const getHistoryQuery = `
SELECT * FROM messages
WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
ORDER BY created_at DESC
LIMIT 100
`

func (m messageRepo) GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.QueryxContext(ctx, getHistoryQuery, params.SenderID, params.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message message
		if err = rows.StructScan(&message); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, models.Message{
			ID:         message.ID,
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Content:    message.Content,
			SentAt:     message.SentAt,
			CreatedAt:  message.CreatedAt,
			UpdatedAt:  message.UpdatedAt,
			DeletedAt:  message.DeletedAt,
		})
	}

	return messages, nil
}
