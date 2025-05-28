package repo

import (
	"context"
	"messanger/internal/models"
)

type RoomRepository interface {
	GetRoomByUsers(ctx context.Context, user1, user2 int64) (*models.Room, error)
	CreateRoom(ctx context.Context, userIDs []int64) (*models.Room, error)
	GetUserIDsInRoom(ctx context.Context, roomID int64) ([]int64, error)
}
