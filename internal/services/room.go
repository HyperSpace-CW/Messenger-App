package services

import (
	"context"
	"fmt"
	"messanger/internal/models"
	"messanger/internal/repo/pg"
)

// RoomService реализует бизнес-логику работы с комнатами
type RoomService interface {
	CreateRoom(ctx context.Context, params CreateRoomParams) (*models.Room, error)
	AddMember(ctx context.Context, params AddMemberParams) error
	RemoveMember(ctx context.Context, params RemoveMemberParams) error
	GetRoom(ctx context.Context, params GetRoomParams) (*models.Room, error)
	GetUserRooms(ctx context.Context, userID int64) ([]models.Room, error)
	UpdateRoom(ctx context.Context, params UpdateRoomParams) error
	DeleteRoom(ctx context.Context, params DeleteRoomParams) error
	GetRoomMembers(ctx context.Context, params GetRoomMembersParams) ([]models.RoomMember, error)
}

type roomService struct {
	roomRepo    pg.RoomRepository
	messageRepo pg.MessageRepo
}

func NewRoomService(roomRepo pg.RoomRepository, messageRepo pg.MessageRepo) RoomService {
	return &roomService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
	}
}

type CreateRoomParams struct {
	Name        string
	Description string
	CreatorID   int64
}

func (s *roomService) CreateRoom(ctx context.Context, params CreateRoomParams) (*models.Room, error) {
	// Валидация параметров
	if params.Name == "" {
		return nil, fmt.Errorf("room name cannot be empty")
	}
	if params.CreatorID == 0 {
		return nil, fmt.Errorf("invalid creator ID")
	}

	// Создание комнаты в репозитории
	roomParams := pg.CreateRoomParams{
		Name:        params.Name,
		Description: params.Description,
		CreatorID:   params.CreatorID,
	}

	if err := s.roomRepo.CreateRoom(ctx, roomParams); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Получаем созданную комнату (в реальной реализации нужно либо возвращать ID из CreateRoom, либо искать по имени)
	// Здесь упрощенная реализация
	rooms, err := s.roomRepo.GetRooms(ctx, params.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created room: %w", err)
	}

	var createdRoom *models.Room
	for _, room := range rooms {
		if room.Name == params.Name && room.CreatorID == params.CreatorID {
			createdRoom = &room
			break
		}
	}

	if createdRoom == nil {
		return nil, fmt.Errorf("failed to find created room")
	}

	return createdRoom, nil
}

type AddMemberParams struct {
	RoomID  int64
	UserID  int64
	IsAdmin bool
}

func (s *roomService) AddMember(ctx context.Context, params AddMemberParams) error {
	// Проверка существования комнаты
	if _, err := s.roomRepo.GetRoomByID(ctx, pg.GetRoomByIDParams{RoomID: params.RoomID}); err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	// Добавление участника
	if err := s.roomRepo.AddMemberToRoom(ctx, pg.AddMemberParams{
		RoomID:  params.RoomID,
		UserID:  params.UserID,
		IsAdmin: params.IsAdmin,
	}); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

type RemoveMemberParams struct {
	RoomID int64
	UserID int64
}

func (s *roomService) RemoveMember(ctx context.Context, params RemoveMemberParams) error {
	// Проверка что пользователь является участником комнаты
	members, err := s.roomRepo.GetRoomMembers(ctx, pg.GetRoomMembersParams{RoomID: params.RoomID})
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	var isMember bool
	for _, member := range members {
		if member.UserID == params.UserID {
			isMember = true
			break
		}
	}

	if !isMember {
		return fmt.Errorf("user is not a member of this room")
	}

	// Удаление участника
	if err := s.roomRepo.RemoveMemberFromRoom(ctx, pg.RemoveMemberParams{
		RoomID: params.RoomID,
		UserID: params.UserID,
	}); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

type GetRoomParams struct {
	RoomID int64
}

func (s *roomService) GetRoom(ctx context.Context, params GetRoomParams) (*models.Room, error) {
	room, err := s.roomRepo.GetRoomByID(ctx, pg.GetRoomByIDParams{RoomID: params.RoomID})
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return room, nil
}

func (s *roomService) GetUserRooms(ctx context.Context, userID int64) ([]models.Room, error) {
	rooms, err := s.roomRepo.GetRooms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}

	return rooms, nil
}

type UpdateRoomParams struct {
	RoomID      int64
	Name        string
	Description string
}

func (s *roomService) UpdateRoom(ctx context.Context, params UpdateRoomParams) error {
	// Проверка существования комнаты
	if _, err := s.roomRepo.GetRoomByID(ctx, pg.GetRoomByIDParams{RoomID: params.RoomID}); err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	// Обновление комнаты
	if err := s.roomRepo.UpdateRoom(ctx, pg.UpdateRoomParams{
		RoomID:      params.RoomID,
		Name:        params.Name,
		Description: params.Description,
	}); err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	return nil
}

type DeleteRoomParams struct {
	RoomID int64
}

func (s *roomService) DeleteRoom(ctx context.Context, params DeleteRoomParams) error {
	// Удаление комнаты
	if err := s.roomRepo.DeleteRoom(ctx, pg.DeleteRoomParams{RoomID: params.RoomID}); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}

type GetRoomMembersParams struct {
	RoomID int64
}

func (s *roomService) GetRoomMembers(ctx context.Context, params GetRoomMembersParams) ([]models.RoomMember, error) {
	members, err := s.roomRepo.GetRoomMembers(ctx, pg.GetRoomMembersParams{RoomID: params.RoomID})
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}

	return members, nil
}
