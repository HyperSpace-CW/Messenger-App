package models

import "time"

type Message struct {
<<<<<<< Updated upstream
	ID         int64
=======
	ID         string
>>>>>>> Stashed changes
	SenderID   int64
	ReceiverID int64
	Content    string
	SentAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
