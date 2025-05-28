package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"messanger/internal/services"
	"messanger/internal/transport/utils"
	"net/http"
	"sync"
)

type WebSocketServer struct {
	messageService services.MessageService
	clients        map[*websocket.Conn]int64
	mu             sync.Mutex
	log            *logrus.Logger
	upgrader       websocket.Upgrader
	tokenKey       string
}

func NewWebSocketServer(messageService services.MessageService, log *logrus.Logger, tokenKey string) *WebSocketServer {
	return &WebSocketServer{
		messageService: messageService,
		clients:        make(map[*websocket.Conn]int64),
		log:            log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Разрешить все соединения
			},
		},
		tokenKey: tokenKey,
	}
}

type CreateMessageRequest struct {
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
}

func (s *WebSocketServer) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Errorf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Извлечение заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		s.log.Error("Authorization header is missing")
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		conn.Close()
		return
	}

	// Извлечение userID из токена
	userID, err := utils.ExtractUserIDFromHeader(authHeader, s.tokenKey)
	if err != nil {
		s.log.Errorf("Failed to extract user ID from header: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		conn.Close()
		return
	}

	s.mu.Lock()
	s.clients[conn] = userID
	s.mu.Unlock()

	s.log.Infof("New client connected: userID=%d", userID)

	for {
		var req CreateMessageRequest
		if err := conn.ReadJSON(&req); err != nil {
			s.log.Infof("conn.ReadJSON: %v", err)
			break
		}

		err := s.messageService.SaveMessage(context.Background(), services.SaveMessageParams{
			SenderID:   userID,
			ReceiverID: req.ReceiverID,
			Content:    req.Content,
		})
		if err != nil {
			s.log.Infof("s.messageService.SaveMessage: %v", err)
			continue
		}

		s.broadcastMessage(userID, req)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()

	s.log.Infof("Client disconnected: userID=%d", userID)
}

func (s *WebSocketServer) broadcastMessage(userID int64, req CreateMessageRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if s.clients[client] != userID {
			if s.clients[client] == req.ReceiverID {
				if err := client.WriteJSON(req.Content); err != nil {
					s.log.Warnf("Error sending message: %v", err)
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *WebSocketServer) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleConnection)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.log.Infof("WebSocket server started on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("http.ListenAndServe: %w", err)
	}
	return nil
}
