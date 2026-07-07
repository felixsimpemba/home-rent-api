package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
	wslib "github.com/felixsimpemba/home-rent-api/internal/ws"
)

// wsUpgrader configures the WebSocket upgrader with permissive origin check.
// In production, replace CheckOrigin with strict domain validation.
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Validate r.Header.Get("Origin") against allowed domains
		return true
	},
}

// WebSocketsHandler handles messaging, notifications and WebSocket connections
type WebSocketsHandler struct {
	db  *gorm.DB
	hub *wslib.Hub
}

// NewWebSocketsHandler creates a WebSocketsHandler with DB and WS hub
func NewWebSocketsHandler(db *gorm.DB, hub *wslib.Hub) *WebSocketsHandler {
	return &WebSocketsHandler{db: db, hub: hub}
}

// ─── WebSocket Connect ────────────────────────────────────────────────────────

// Connect upgrades the HTTP connection to WebSocket and registers the client
// with the hub for real-time message delivery.
//
// Expected client message frame:
//
//	{ "event": "message", "payload": { "conversation_id": "...", "content": "..." } }
//	{ "event": "typing",  "payload": { "conversation_id": "..." } }
func (h *WebSocketsHandler) Connect(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// Upgrade to WebSocket
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own HTTP error; just log
		return
	}

	// Create client and register with hub
	client := wslib.NewClient(h.hub, userID, conn)
	h.hub.Register(userID, client)

	// Send a welcome message immediately
	h.hub.Send(userID, wslib.EventPing, map[string]interface{}{
		"message": "Connected to HomeRent real-time service.",
		"user_id": userID,
	})

	// Start read/write pumps — this is non-blocking; goroutines take over
	client.Run(h.handleIncoming)
}

// handleIncoming processes frames sent from the client to the server
func (h *WebSocketsHandler) handleIncoming(senderID string, frame wslib.IncomingFrame) {
	switch frame.Event {

	case wslib.EventMessage:
		var payload struct {
			ConversationID string `json:"conversation_id"`
			Content        string `json:"content"`
		}
		if err := json.Unmarshal(frame.Payload, &payload); err != nil || payload.Content == "" {
			return
		}
		h.deliverMessage(senderID, payload.ConversationID, payload.Content)

	case wslib.EventTyping:
		var payload struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return
		}
		h.broadcastTyping(senderID, payload.ConversationID)

	case wslib.EventReadReceipt:
		var payload struct {
			MessageID string `json:"message_id"`
		}
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return
		}
		// Notify the message sender that it was read
		// (not persisted for now — ephemeral signal)
		var msg models.Message
		if h.db.First(&msg, "id = ?", payload.MessageID).Error == nil {
			h.hub.Send(msg.SenderID, wslib.EventReadReceipt, map[string]interface{}{
				"message_id": payload.MessageID,
				"read_by":    senderID,
			})
		}

	case wslib.EventPing:
		h.hub.Send(senderID, wslib.EventPing, map[string]interface{}{"pong": true})
	}
}

// deliverMessage persists a message and pushes it to the recipient via WS (or stores it for polling)
func (h *WebSocketsHandler) deliverMessage(senderID, conversationID, content string) {
	// Verify sender belongs to the conversation
	var conv models.Conversation
	if err := h.db.First(&conv, "id = ?", conversationID).Error; err != nil {
		return
	}
	if conv.Participant1 != senderID && conv.Participant2 != senderID {
		return // not a participant
	}

	// Persist message
	msg := models.Message{
		ID:             "msg_" + uuid.NewString(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
	}
	if err := h.db.Create(&msg).Error; err != nil {
		return
	}

	// Update conversation last message
	h.db.Model(&conv).Updates(map[string]interface{}{
		"last_message": content,
	})

	// Determine recipient
	recipientID := conv.Participant1
	if senderID == conv.Participant1 {
		recipientID = conv.Participant2
	}

	// Preload sender details for the response payload
	h.db.Preload("Sender").First(&msg, "id = ?", msg.ID)

	msgPayload := map[string]interface{}{
		"id":              msg.ID,
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"sender": map[string]interface{}{
			"id":         msg.Sender.ID,
			"first_name": msg.Sender.FirstName,
			"last_name":  msg.Sender.LastName,
		},
		"content":    content,
		"created_at": msg.CreatedAt,
	}

	// Push to recipient (if online)
	online := h.hub.Send(recipientID, wslib.EventMessage, msgPayload)

	// Also echo to sender so their UI can confirm delivery
	h.hub.Send(senderID, wslib.EventMessage, msgPayload)

	// If recipient is offline, persist a notification
	if !online {
		h.db.Create(&models.Notification{
			ID:          "notif_" + uuid.NewString(),
			RecipientID: recipientID,
			Title:       "New message from " + msg.Sender.FirstName,
			Body:        content,
			Type:        "message",
		})
	}
}

// broadcastTyping sends a typing indicator to the other conversation participant
func (h *WebSocketsHandler) broadcastTyping(senderID, conversationID string) {
	var conv models.Conversation
	if err := h.db.First(&conv, "id = ?", conversationID).Error; err != nil {
		return
	}

	recipientID := conv.Participant1
	if senderID == conv.Participant1 {
		recipientID = conv.Participant2
	}

	h.hub.Send(recipientID, wslib.EventTyping, map[string]interface{}{
		"conversation_id": conversationID,
		"user_id":         senderID,
	})
}

// ─── Conversations ────────────────────────────────────────────────────────────

func (h *WebSocketsHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Conversation{}).
		Where("participant1 = ? OR participant2 = ?", userID, userID).
		Count(&total)

	var convs []models.Conversation
	h.db.Where("participant1 = ? OR participant2 = ?", userID, userID).
		Offset((page - 1) * limit).Limit(limit).
		Order("updated_at DESC").Find(&convs)

	// Annotate with online status for each participant
	type ConvWithStatus struct {
		models.Conversation
		Participant1Online bool `json:"participant1_online"`
		Participant2Online bool `json:"participant2_online"`
	}
	result := make([]ConvWithStatus, 0, len(convs))
	for _, c := range convs {
		result = append(result, ConvWithStatus{
			Conversation:       c,
			Participant1Online: h.hub.IsOnline(c.Participant1),
			Participant2Online: h.hub.IsOnline(c.Participant2),
		})
	}

	respond.Paginated(w, result, page, limit, total)
}

func (h *WebSocketsHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		ParticipantID string `json:"participant_id"`
		PropertyID    string `json:"property_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ParticipantID == "" {
		apierrors.BadRequest(w, r, "participant_id is required.", nil)
		return
	}

	// Return existing conversation if one already exists
	var existing models.Conversation
	err := h.db.Where(
		"(participant1 = ? AND participant2 = ?) OR (participant1 = ? AND participant2 = ?)",
		userID, body.ParticipantID, body.ParticipantID, userID,
	).First(&existing).Error

	if err == nil {
		respond.OK(w, existing)
		return
	}

	conv := models.Conversation{
		ID:           "conv_" + uuid.NewString(),
		Participant1: userID,
		Participant2: body.ParticipantID,
		PropertyID:   body.PropertyID,
	}
	h.db.Create(&conv)
	respond.Created201(w, conv)
}

func (h *WebSocketsHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := r.PathValue("id")

	var conv models.Conversation
	if err := h.db.First(&conv, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Conversation not found.")
		return
	}

	if conv.Participant1 != userID && conv.Participant2 != userID {
		apierrors.Forbidden(w, r, "Access denied.")
		return
	}

	respond.OK(w, map[string]interface{}{
		"conversation":        conv,
		"participant1_online": h.hub.IsOnline(conv.Participant1),
		"participant2_online": h.hub.IsOnline(conv.Participant2),
	})
}

// ─── HTTP Messages ────────────────────────────────────────────────────────────

func (h *WebSocketsHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Message{}).Where("conversation_id = ?", id).Count(&total)

	var messages []models.Message
	h.db.Where("conversation_id = ?", id).
		Preload("Sender").
		Offset((page - 1) * limit).Limit(limit).
		Order("created_at ASC").Find(&messages)

	respond.Paginated(w, messages, page, limit, total)
}

// SendMessageHTTP sends a message over HTTP and delivers it via WebSocket if the recipient is online.
func (h *WebSocketsHandler) SendMessageHTTP(w http.ResponseWriter, r *http.Request) {
	senderID := middleware.GetUserID(r)
	convID := r.PathValue("id")

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		apierrors.BadRequest(w, r, "content is required.", nil)
		return
	}

	// Delegate to the same delivery logic used by WebSocket frames
	h.deliverMessage(senderID, convID, body.Content)

	// Return the created message from DB
	var msg models.Message
	h.db.Preload("Sender").Where("conversation_id = ? AND sender_id = ?", convID, senderID).
		Last(&msg)

	respond.Created201(w, msg)
}

func (h *WebSocketsHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	msgID := r.PathValue("msgId")

	var msg models.Message
	if err := h.db.First(&msg, "id = ?", msgID).Error; err != nil {
		apierrors.NotFound(w, r, "Message not found.")
		return
	}
	if msg.SenderID != userID {
		apierrors.Forbidden(w, r, "You can only delete your own messages.")
		return
	}
	h.db.Delete(&msg)
	respond.NoContent(w)
}

// ─── Notifications ────────────────────────────────────────────────────────────

func (h *WebSocketsHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Notification{}).Where("recipient_id = ?", userID).Count(&total)

	var notifications []models.Notification
	h.db.Where("recipient_id = ?", userID).
		Offset((page - 1) * limit).Limit(limit).
		Order("created_at DESC").Find(&notifications)

	respond.Paginated(w, notifications, page, limit, total)
}

func (h *WebSocketsHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Model(&models.Notification{}).
		Where("id = ? AND recipient_id = ?", id, userID).
		Update("read", true)
	respond.Message(w, "Notification marked as read.")
}

func (h *WebSocketsHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	h.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND read = false", userID).
		Update("read", true)
	respond.Message(w, "All notifications marked as read.")
}

func (h *WebSocketsHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Where("id = ? AND recipient_id = ?", id, userID).Delete(&models.Notification{})
	respond.NoContent(w)
}
