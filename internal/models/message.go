package models

import "time"

type Message struct {
	ID         int64
	SenderID   int64
	ReceiverID int64
	Content    string
	SentAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
