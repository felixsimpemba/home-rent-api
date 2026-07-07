// Package ws implements the real-time WebSocket hub for HomeRent.
//
// Architecture:
//   - Hub: singleton that owns a map[userID]*Client
//   - Client: wraps a gorilla WebSocket conn with separate read/write goroutines
//   - Messages flow: HTTP POST or WS frame → Hub.Send(recipientID, payload) → Client.send channel → write pump
package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// ─── Event types sent to clients ──────────────────────────────────────────────

const (
	EventMessage      = "message"       // new chat message
	EventNotification = "notification"  // in-app notification
	EventTyping       = "typing"        // typing indicator
	EventReadReceipt  = "read_receipt"  // message read confirmation
	EventPing         = "ping"          // keepalive
)

// Envelope is the standard envelope for all WS events
type Envelope struct {
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewEnvelope wraps any payload into an Envelope
func NewEnvelope(event string, payload interface{}) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{
		Event:     event,
		Payload:   raw,
		Timestamp: time.Now(),
	})
}

// ─── Hub ──────────────────────────────────────────────────────────────────────

// Hub maintains the set of active WebSocket clients and broadcasts messages.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // userID → *Client
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

// Register adds a client to the hub
func (h *Hub) Register(userID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Close existing connection for the same user (single-session model)
	if existing, ok := h.clients[userID]; ok {
		existing.close()
	}
	h.clients[userID] = c
	log.Printf("[WS] User %s connected (total online: %d)", userID, len(h.clients))
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		c.close()
		delete(h.clients, userID)
	}
	log.Printf("[WS] User %s disconnected (total online: %d)", userID, len(h.clients))
}

// Send delivers a typed event payload to a specific user.
// Returns true if the user was online and the message was queued.
func (h *Hub) Send(userID string, event string, payload interface{}) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return false // user not connected
	}

	data, err := NewEnvelope(event, payload)
	if err != nil {
		log.Printf("[WS] Failed to encode envelope for user %s: %v", userID, err)
		return false
	}

	select {
	case client.send <- data:
		return true
	default:
		// Channel full — client is too slow, drop and disconnect
		h.Unregister(userID)
		return false
	}
}

// Broadcast delivers a message to multiple users
func (h *Hub) Broadcast(userIDs []string, event string, payload interface{}) {
	for _, uid := range userIDs {
		h.Send(uid, event, payload)
	}
}

// OnlineUsers returns the list of currently connected user IDs
func (h *Hub) OnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

// IsOnline returns true if the given userID has an active connection
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
