package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"tuno_backend/internal/domain"
	"tuno_backend/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type IncomingMessage struct {
	GroupID string `json:"group_id"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type SendMessageCommand struct {
	UserID  string
	GroupID string
	Content string
	Type    domain.MessageType
}

type SendMessageResult struct {
	Message   interface{}
	MemberIDs []string
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now, should be restricted in production
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// User ID
	UserID string

	// Command Bus
	commandBus CommandBus

	// Buffered channel of outbound messages.
	send chan []byte
}

type CommandBus interface {
	Dispatch(ctx context.Context, cmd interface{}) (interface{}, error)
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				logger.Error("Unexpected websocket close error", zap.Error(err))
			}
			break
		}

		// Parse incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			logger.Error("Failed to parse incoming message", zap.Error(err))
			continue
		}

		// Create Command
		cmd := SendMessageCommand{
			UserID:  c.UserID,
			GroupID: incomingMsg.GroupID,
			Content: incomingMsg.Content,
			Type:    domain.MessageType(incomingMsg.Type),
		}

		// Dispatch Command
		// Use a timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := c.commandBus.Dispatch(ctx, cmd)
		cancel()

		if err != nil {
			logger.Error("Failed to dispatch send message command", zap.Error(err))
			// TODO: Send error back to client?
			continue
		}

		// Broadcast Result
		sendResult, ok := result.(SendMessageResult)
		if !ok {
			logger.Error("Command result is not SendMessageResult")
			continue
		}

		// Serialize Message
		msgBytes, err := json.Marshal(sendResult.Message)
		if err != nil {
			logger.Error("Failed to marshal message", zap.Error(err))
			continue
		}

		// Fan-out via Hub
		c.hub.BroadcastToUsers(sendResult.MemberIDs, msgBytes)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID string, commandBus CommandBus) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade websocket", zap.Error(err))
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), UserID: userID, commandBus: commandBus}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
