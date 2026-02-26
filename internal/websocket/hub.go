package websocket

import (
	"tuno_backend/pkg/logger"

	"go.uber.org/zap"
)

type GroupMessage struct {
	GroupID string
	Payload []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Clients mapped by UserID
	clientsByUserID map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Inbound messages targeted to specific group
	groupBroadcast chan GroupMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:       make(chan []byte),
		groupBroadcast:  make(chan GroupMessage),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[*Client]bool),
		clientsByUserID: make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if client.UserID != "" {
				h.clientsByUserID[client.UserID] = client
			}
			logger.Info("Client registered", zap.String("addr", client.conn.RemoteAddr().String()), zap.String("user_id", client.UserID))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.UserID != "" {
					delete(h.clientsByUserID, client.UserID)
				}
				close(client.send)
				logger.Info("Client unregistered", zap.String("addr", client.conn.RemoteAddr().String()))
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					if client.UserID != "" {
						delete(h.clientsByUserID, client.UserID)
					}
				}
			}
		case groupMsg := <-h.groupBroadcast:
			// No-op for now as we don't have direct GroupID -> Clients mapping yet.
			// Use BroadcastToUsers from outside.
			_ = groupMsg
		}
	}
}

func (h *Hub) BroadcastToUsers(userIDs []string, payload []byte) {
	for _, userID := range userIDs {
		if client, ok := h.clientsByUserID[userID]; ok {
			select {
			case client.send <- payload:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(h.clientsByUserID, userID)
			}
		}
	}
}
