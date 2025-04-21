package repo

import (
	"context"
	"fmt"
	"time"

	"messanger/config"
	"messanger/internal/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type MessageRepo interface {
	SaveMessage(ctx context.Context, params SaveMessageParams) error
	GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error)
}

type messageRepo struct {
	db *sqlx.DB
}

func NewMessageRepo(cfg *config.Config) MessageRepo {
	db := sqlx.MustConnect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PG.Host,
		cfg.PG.Port,
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.DBName,
		cfg.PG.SSLMode,
	))

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := Migration(db); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	return &messageRepo{db: db}
}

func Migration(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repo/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

type message struct {
	ID         int64      `db:"id"`
	SenderID   int64      `db:"sender_id"`
	ReceiverID int64      `db:"receiver_id"`
	Content    string     `db:"content"`
	SentAt     *time.Time `db:"sent_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

type SaveMessageParams struct {
	SenderID   int64
	ReceiverID int64
	Content    string
}

const saveMessageQuery = `
INSERT INTO messages (sender_id, receiver_id, content)
VALUES ($1, $2, $3)
`

func (m messageRepo) SaveMessage(ctx context.Context, params SaveMessageParams) error {
	_, err := m.db.ExecContext(ctx, saveMessageQuery, params.SenderID, params.ReceiverID, params.Content)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

type GetHistoryParams struct {
	SenderID   int64
	ReceiverID int64
}

const getHistoryQuery = `
SELECT * FROM messages
WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
ORDER BY created_at DESC
LIMIT 100
`

func (m messageRepo) GetHistory(ctx context.Context, params GetHistoryParams) ([]models.Message, error) {
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
