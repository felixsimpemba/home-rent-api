package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192 // 8KB
)

// IncomingFrame is the structure of messages sent by clients to the server
type IncomingFrame struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

// Client represents a connected WebSocket peer
type Client struct {
	hub    *Hub
	userID string
	conn   *websocket.Conn
	send   chan []byte // buffered channel of outbound messages
}

// NewClient creates a new Client
func NewClient(hub *Hub, userID string, conn *websocket.Conn) *Client {
	return &Client{
		hub:    hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

// close shuts down the send channel (called by hub, safe to call once)
func (c *Client) close() {
	select {
	case <-c.send:
	default:
		close(c.send)
	}
}

// Run starts both the read and write pumps in separate goroutines
func (c *Client) Run(onIncoming func(userID string, frame IncomingFrame)) {
	go c.writePump()
	go c.readPump(onIncoming)
}

// ─── Read Pump ────────────────────────────────────────────────────────────────

// readPump reads frames sent by the browser to the server.
// It runs in its own goroutine per connection.
func (c *Client) readPump(onIncoming func(userID string, frame IncomingFrame)) {
	defer func() {
		c.hub.Unregister(c.userID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Read error for user %s: %v", c.userID, err)
			}
			break
		}

		var frame IncomingFrame
		if err := json.Unmarshal(raw, &frame); err != nil {
			log.Printf("[WS] Invalid frame from user %s: %v", c.userID, err)
			continue
		}

		if onIncoming != nil {
			onIncoming(c.userID, frame)
		}
	}
}

// ─── Write Pump ───────────────────────────────────────────────────────────────

// writePump pushes messages from the hub to the WebSocket connection.
// It runs in its own goroutine per connection.
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Flush any queued messages in the same write frame
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
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
