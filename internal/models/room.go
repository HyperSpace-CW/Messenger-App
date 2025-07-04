package models

import "time"

type Room struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	CreatorID   int64     `json:"creator_id"`
}

type RoomMember struct {
	RoomID   int64     `json:"room_id"`
	UserID   int64     `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	IsAdmin  bool      `json:"is_admin"`
}

type UserRoom struct {
	Room        Room     `json:"room"`
	UnreadCount int      `json:"unread_count"`
	LastMessage *Message `json:"last_message,omitempty"`
}
