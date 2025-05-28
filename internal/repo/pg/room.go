package pg

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"messanger/internal/models"
	"sync"
	"time"
)

type RoomRepository interface {
	GetRooms(ctx context.Context, userID int64) ([]models.Room, error)
	CreateRoom(ctx context.Context, params CreateRoomParams) error
	AddMemberToRoom(ctx context.Context, params AddMemberParams) error
	GetRoomByID(ctx context.Context, params GetRoomByIDParams) (*models.Room, error)
	RemoveMemberFromRoom(ctx context.Context, params RemoveMemberParams) error
	UpdateRoom(ctx context.Context, params UpdateRoomParams) error
	DeleteRoom(ctx context.Context, params DeleteRoomParams) error
	GetRoomMembers(ctx context.Context, params GetRoomMembersParams) ([]models.RoomMember, error)
	GetMessages(ctx context.Context, params GetMessagesParams) ([]models.Message, error)
	CreateMessage(ctx context.Context, params CreateMessageParams) error
	UpdateMessage(ctx context.Context, params UpdateMessageParams) error
	DeleteMessage(ctx context.Context, params DeleteMessageParams) error
	GetMessageByID(ctx context.Context, params GetMessageByIDParams) (*models.Message, error)
}

type roomRepository struct {
	db *sqlx.DB
	mu *sync.RWMutex
}

func NewRoomRepository(db *sqlx.DB) RoomRepository {
	return &roomRepository{
		db: db,
		mu: new(sync.RWMutex),
	}
}

type room struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	CreatorID   int64     `db:"creator_id"`
}

type roomMember struct {
	RoomID   int64     `db:"room_id"`
	UserID   int64     `db:"user_id"`
	JoinedAt time.Time `db:"joined_at"`
	IsAdmin  bool      `db:"is_admin"`
}

// Структуры параметров
type CreateRoomParams struct {
	Name        string
	Description string
	CreatorID   int64
}

type AddMemberParams struct {
	RoomID  int64
	UserID  int64
	IsAdmin bool
}

type RemoveMemberParams struct {
	RoomID int64
	UserID int64
}

type GetRoomByIDParams struct {
	RoomID int64
}

type UpdateRoomParams struct {
	RoomID      int64
	Name        string
	Description string
}

type DeleteRoomParams struct {
	RoomID int64
}

type GetRoomMembersParams struct {
	RoomID int64
}

type GetMessagesParams struct {
	RoomID    int64
	Limit     int
	Offset    int
	StartTime time.Time
	EndTime   time.Time
}

type CreateMessageParams struct {
	RoomID   int64
	SenderID int64
	Content  string
}

type UpdateMessageParams struct {
	MessageID int64
	Content   string
}

type DeleteMessageParams struct {
	MessageID int64
}

type GetMessageByIDParams struct {
	MessageID int64
}

// Реализации методов

const getRoomsQuery = `
SELECT r.id, r.name, r.description, r.created_at, r.creator_id 
FROM rooms r
JOIN room_members rm ON r.id = rm.room_id
WHERE rm.user_id = $1
`

func (r *roomRepository) GetRooms(ctx context.Context, userID int64) ([]models.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var rooms []room
	err := r.db.SelectContext(ctx, &rooms, getRoomsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	result := make([]models.Room, len(rooms))
	for i, ro := range rooms {
		result[i] = models.Room{
			ID:          ro.ID,
			Name:        ro.Name,
			Description: ro.Description,
			CreatedAt:   ro.CreatedAt,
			CreatorID:   ro.CreatorID,
		}
	}

	return result, nil
}

const createRoomQuery = `
INSERT INTO rooms (id, name, description, created_at, creator_id)
VALUES ($1, $2, $3, $4, $5)
`

func (r *roomRepository) CreateRoom(ctx context.Context, params CreateRoomParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(
		ctx,
		createRoomQuery,
		uuid.New().ID(),
		params.Name,
		params.Description,
		time.Now(),
		params.CreatorID,
	)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const addMemberQuery = `
INSERT INTO room_members (room_id, user_id, is_admin, joined_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (room_id, user_id) DO NOTHING
`

func (r *roomRepository) AddMemberToRoom(ctx context.Context, params AddMemberParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(
		ctx,
		addMemberQuery,
		params.RoomID,
		params.UserID,
		params.IsAdmin,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const getRoomByIDQuery = `
SELECT id, name, description, created_at, creator_id 
FROM rooms 
WHERE id = $1
`

func (r *roomRepository) GetRoomByID(ctx context.Context, params GetRoomByIDParams) (*models.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ro room
	err := r.db.GetContext(ctx, &ro, getRoomByIDQuery, params.RoomID)
	if err != nil {
		return nil, fmt.Errorf("r.db.GetContext: %w", err)
	}

	return &models.Room{
		ID:          ro.ID,
		Name:        ro.Name,
		Description: ro.Description,
		CreatedAt:   ro.CreatedAt,
		CreatorID:   ro.CreatorID,
	}, nil
}

const removeMemberQuery = `
DELETE FROM room_members 
WHERE room_id = $1 AND user_id = $2
`

func (r *roomRepository) RemoveMemberFromRoom(ctx context.Context, params RemoveMemberParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(ctx, removeMemberQuery, params.RoomID, params.UserID)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const updateRoomQuery = `
UPDATE rooms 
SET name = $2, description = $3 
WHERE id = $1
`

func (r *roomRepository) UpdateRoom(ctx context.Context, params UpdateRoomParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(ctx, updateRoomQuery, params.RoomID, params.Name, params.Description)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const deleteRoomQuery = `
DELETE FROM rooms 
WHERE id = $1
`

func (r *roomRepository) DeleteRoom(ctx context.Context, params DeleteRoomParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(ctx, deleteRoomQuery, params.RoomID)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const getRoomMembersQuery = `
SELECT room_id, user_id, joined_at, is_admin 
FROM room_members 
WHERE room_id = $1
`

func (r *roomRepository) GetRoomMembers(ctx context.Context, params GetRoomMembersParams) ([]models.RoomMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var members []roomMember
	err := r.db.SelectContext(ctx, &members, getRoomMembersQuery, params.RoomID)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	result := make([]models.RoomMember, len(members))
	for i, m := range members {
		result[i] = models.RoomMember{
			RoomID:   m.RoomID,
			UserID:   m.UserID,
			JoinedAt: m.JoinedAt,
			IsAdmin:  m.IsAdmin,
		}
	}

	return result, nil
}

const getMessagesQuery = `
SELECT id, sender_id, receiver_id, content, sent_at, created_at, updated_at, deleted_at, error_message 
FROM messages 
WHERE room_id = $1 
AND created_at BETWEEN $2 AND $3 
ORDER BY created_at DESC 
LIMIT $4 OFFSET $5
`

func (r *roomRepository) GetMessages(ctx context.Context, params GetMessagesParams) ([]models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []models.Message
	err := r.db.SelectContext(ctx, &messages, getMessagesQuery,
		params.RoomID,
		params.StartTime,
		params.EndTime,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	return messages, nil
}

const createMessageQuery = `
INSERT INTO messages (room_id, sender_id, receiver_id, content, sent_at)
VALUES ($1, $2, $3, $4, $5)
`

func (r *roomRepository) CreateMessage(ctx context.Context, params CreateMessageParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO
	// В реальном приложении нужно определить получателя
	// Здесь для примера просто используем 0
	receiverID := int64(0)

	_, err := r.db.ExecContext(
		ctx,
		createMessageQuery,
		params.RoomID,
		params.SenderID,
		receiverID,
		params.Content,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const updateMessageQuery = `
UPDATE messages 
SET content = $2, updated_at = NOW() 
WHERE id = $1
`

func (r *roomRepository) UpdateMessage(ctx context.Context, params UpdateMessageParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(ctx, updateMessageQuery, params.MessageID, params.Content)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const deleteMessageQuery = `
UPDATE messages 
SET deleted_at = NOW() 
WHERE id = $1
`

func (r *roomRepository) DeleteMessage(ctx context.Context, params DeleteMessageParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.ExecContext(ctx, deleteMessageQuery, params.MessageID)
	if err != nil {
		return fmt.Errorf("r.db.ExecContext: %w", err)
	}

	return nil
}

const getMessageByIDQuery = `
SELECT id, sender_id, receiver_id, content, sent_at, created_at, updated_at, deleted_at, error_message 
FROM messages 
WHERE id = $1
`

func (r *roomRepository) GetMessageByID(ctx context.Context, params GetMessageByIDParams) (*models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var msg models.Message
	err := r.db.GetContext(ctx, &msg, getMessageByIDQuery, params.MessageID)
	if err != nil {
		return nil, fmt.Errorf("r.db.GetContext: %w", err)
	}

	return &msg, nil
}
