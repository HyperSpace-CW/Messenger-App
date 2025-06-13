package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"messanger/internal/models"
	"messanger/internal/services"
	"messanger/internal/transport/http/v1"
)

// MockMessageService реализует интерфейс services.MessageService для тестов
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SaveMessage(ctx context.Context, params services.SaveMessageParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockMessageService) GetHistory(ctx context.Context, params services.GetHistoryParams) ([]models.Message, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]models.Message), args.Error(1)
}

func TestHandler_getMessagesByID(t *testing.T) {
	type mockBehavior func(s *MockMessageService, receiverID int64)

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
	}

	tests := []struct {
		name             string
		receiverID       string
		mockBehavior     mockBehavior
		expectedStatus   int
		expectedResponse string
		expectedError    string
	}{
		{
			name:       "success",
			receiverID: "2",
			mockBehavior: func(s *MockMessageService, receiverID int64) {
				s.On("GetHistory", mock.Anything, services.GetHistoryParams{
					SenderID:   0,
					ReceiverID: 2,
				}).Return(testMessages, nil)
			},
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `[{"ID":1,"SenderID":1,"ReceiverID":2,"Content":"Hello","SentAt":"` + now.Format(time.RFC3339Nano) + `","CreatedAt":"` + now.Format(time.RFC3339Nano) + `","UpdatedAt":"` + now.Format(time.RFC3339Nano) + `","DeletedAt":null}]`,
		},
		{
			name:           "empty receiverID",
			receiverID:     "",
			mockBehavior:   func(s *MockMessageService, receiverID int64) {},
			expectedStatus: fiber.StatusNotFound,
			//expectedError:  "receiverID is required",
		},
		{
			name:       "no messages found",
			receiverID: "2",
			mockBehavior: func(s *MockMessageService, receiverID int64) {
				s.On("GetHistory", mock.Anything, services.GetHistoryParams{
					SenderID:   0,
					ReceiverID: 2,
				}).Return([]models.Message{}, nil)
			},
			expectedStatus: fiber.StatusNotFound,
			//expectedError:  "no messages found",
		},
		{
			name:       "service error",
			receiverID: "2",
			mockBehavior: func(s *MockMessageService, receiverID int64) {
				s.On("GetHistory", mock.Anything, services.GetHistoryParams{
					SenderID:   0,
					ReceiverID: 2,
				}).Return([]models.Message{}, errors.New("database error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			//expectedError:  "h.messageService.GetHistory: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализация мока сервиса
			messageService := new(MockMessageService)
			receiverIDInt, _ := strconv.ParseInt(tt.receiverID, 10, 64)
			tt.mockBehavior(messageService, receiverIDInt)

			// Создание Fiber приложения и обработчика
			app := fiber.New()
			h := v1.NewHandler(v1.HandlerConfig{
				MessageService: messageService,
				Log:            nil,
			}) // предполагается, что у вас есть конструктор
			app.Get("/messages/:id", h.GetMessagesByID)

			// Создание запроса
			req := httptest.NewRequest("GET", "/messages/"+tt.receiverID, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			if tt.expectedResponse != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.expectedResponse, string(body))
			}

			messageService.AssertExpectations(t)
		})
	}
}

func TestHandler_createMessage(t *testing.T) {
	type mockBehavior func(s *MockMessageService, req v1.CreateMessageRequest)
	type request struct {
		body string
	}

	tests := []struct {
		name           string
		input          request
		mockBehavior   mockBehavior
		expectedStatus int
		expectedError  string
	}{
		{
			name: "success",
			input: request{
				body: `{"sender_id":1,"receiver_id":2,"content":"Hello"}`,
			},
			mockBehavior: func(s *MockMessageService, req v1.CreateMessageRequest) {
				s.On("SaveMessage", mock.Anything, services.SaveMessageParams{
					SenderID:   1,
					ReceiverID: 2,
					Content:    "Hello",
				}).Return(nil)
			},
			expectedStatus: fiber.StatusCreated,
		},
		{
			name: "invalid request body",
			input: request{
				body: `{"sender_id":"1","receiver_id":2,"content":"Hello"}`,
			},
			mockBehavior:   func(s *MockMessageService, req v1.CreateMessageRequest) {},
			expectedStatus: fiber.StatusBadRequest,
			//expectedError:  "failed to parse request body",
		},
		{
			name: "service error",
			input: request{
				body: `{"sender_id":1,"receiver_id":2,"content":"Hello"}`,
			},
			mockBehavior: func(s *MockMessageService, req v1.CreateMessageRequest) {
				s.On("SaveMessage", mock.Anything, services.SaveMessageParams{
					SenderID:   1,
					ReceiverID: 2,
					Content:    "Hello",
				}).Return(errors.New("database error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			//expectedError:  "h.messageService.SaveMessage: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализация мока сервиса
			messageService := new(MockMessageService)

			// Парсим тело запроса для передачи в mockBehavior
			var req v1.CreateMessageRequest
			if tt.input.body != "" {
				err := json.Unmarshal([]byte(tt.input.body), &req)
				if err == nil {
					tt.mockBehavior(messageService, req)
				}
			}

			// Создание Fiber приложения и обработчика
			app := fiber.New()
			h := v1.NewHandler(v1.HandlerConfig{
				MessageService: messageService,
				Log:            nil,
			}) // предполагается, что у вас есть конструктор
			app.Post("/messages", h.CreateMessage)

			// Создание запроса
			reqHTTP := httptest.NewRequest("POST", "/messages", bytes.NewBufferString(tt.input.body))
			reqHTTP.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(reqHTTP)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var response map[string]string
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			messageService.AssertExpectations(t)
		})
	}
}
