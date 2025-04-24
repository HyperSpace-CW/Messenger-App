package models

import "time"

type Message struct {
	ID         string
	SenderID   string
	ReceiverID string
	Content    string
	SentAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
