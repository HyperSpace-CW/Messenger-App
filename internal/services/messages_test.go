package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"messanger/internal/models"
	"messanger/internal/repo"
	"messanger/internal/services"
)

// MockMessageRepo реализует интерфейс repo.MessageRepo для тестов
type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) SaveMessage(ctx context.Context, params repo.SaveMessageParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockMessageRepo) GetHistory(ctx context.Context, params repo.GetHistoryParams) ([]models.Message, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]models.Message), args.Error(1)
}

func TestMessageService_SaveMessage(t *testing.T) {
	tests := []struct {
		name          string
		params        services.SaveMessageParams
		repoSetup     func(*MockMessageRepo)
		expectedError error
	}{
		{
			name: "successful message save",
			params: services.SaveMessageParams{
				SenderID:   1,
				ReceiverID: 2,
				Content:    "Hello",
			},
			repoSetup: func(m *MockMessageRepo) {
				m.On("SaveMessage", mock.Anything, repo.SaveMessageParams{
					SenderID:   1,
					ReceiverID: 2,
					Content:    "Hello",
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			params: services.SaveMessageParams{
				SenderID:   1,
				ReceiverID: 2,
				Content:    "Hello",
			},
			repoSetup: func(m *MockMessageRepo) {
				m.On("SaveMessage", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			expectedError: errors.New("s.repo.SaveMessage: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMessageRepo)
			tt.repoSetup(mockRepo)

			service := services.NewMessageService(mockRepo)
			err := service.SaveMessage(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_GetHistory(t *testing.T) {
	now := time.Now()
	testMessages := []models.Message{
		{
			ID:         1,
			SenderID:   1,
			ReceiverID: 2,
			Content:    "Hello",
			SentAt:     &now,
			CreatedAt:  now,
			UpdatedAt:  now,
			DeletedAt:  nil,
		},
		{
			ID:         2,
			SenderID:   2,
			ReceiverID: 1,
			Content:    "Hi there",
			SentAt:     &now,
			CreatedAt:  now.Add(time.Minute),
			UpdatedAt:  now.Add(time.Minute),
			DeletedAt:  nil,
		},
	}

	tests := []struct {
		name           string
		params         services.GetHistoryParams
		repoSetup      func(*MockMessageRepo)
		expectedResult []models.Message
		expectedError  error
	}{
		{
			name: "successful history retrieval",
			params: services.GetHistoryParams{
				SenderID:   1,
				ReceiverID: 2,
			},
			repoSetup: func(m *MockMessageRepo) {
				m.On("GetHistory", mock.Anything, repo.GetHistoryParams{
					SenderID:   1,
					ReceiverID: 2,
				}).Return(testMessages, nil)
			},
			expectedResult: testMessages,
			expectedError:  nil,
		},
		{
			name: "repository error",
			params: services.GetHistoryParams{
				SenderID:   1,
				ReceiverID: 2,
			},
			repoSetup: func(m *MockMessageRepo) {
				m.On("GetHistory", mock.Anything, mock.Anything).
					Return([]models.Message{}, errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("s.repo.GetHistory: database error"),
		},
		{
			name: "empty history",
			params: services.GetHistoryParams{
				SenderID:   3,
				ReceiverID: 4,
			},
			repoSetup: func(m *MockMessageRepo) {
				m.On("GetHistory", mock.Anything, repo.GetHistoryParams{
					SenderID:   3,
					ReceiverID: 4,
				}).Return([]models.Message{}, nil)
			},
			expectedResult: []models.Message{},
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMessageRepo)
			tt.repoSetup(mockRepo)

			service := services.NewMessageService(mockRepo)
			result, err := service.GetHistory(context.Background(), tt.params)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
