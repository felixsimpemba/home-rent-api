package handlers

import (
	"encoding/json"
	"net/http"
)

type WebSocketsHandler struct{}

func NewWebSocketsHandler() *WebSocketsHandler {
	return &WebSocketsHandler{}
}

// Connect upgrades HTTP to WebSocket
func (h *WebSocketsHandler) Connect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver does not support connection hijacking.", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Write mock 101 Switching Protocols response
	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"\r\n")
	bufrw.Flush()

	// Read and discard frames to maintain open socket or mock communication
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
	}
}

// Conversations
func (h *WebSocketsHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	convs := []map[string]interface{}{
		{
			"id": "conv_123",
			"participants": []map[string]interface{}{
				{"id": "usr_tenant_123", "name": "Felix Simpemba"},
				{"id": "usr_landlord_456", "name": "Jane Landlord"},
			},
			"last_message": map[string]interface{}{
				"id":              "msg_901",
				"conversation_id": "conv_123",
				"sender_id":       "usr_tenant_123",
				"content":         "Hi, can I check this house tomorrow?",
				"created_at":      "2026-06-08T17:55:00Z",
			},
			"updated_at": "2026-06-08T17:55:00Z",
		},
	}
	json.NewEncoder(w).Encode(convs)
}

func (h *WebSocketsHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *WebSocketsHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"updated_at": "2026-06-08T17:55:00Z",
	})
}

func (h *WebSocketsHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	messages := []map[string]interface{}{
		{
			"id":              "msg_901",
			"conversation_id": "conv_123",
			"sender_id":       "usr_tenant_123",
			"content":         "Hi, can I check this house tomorrow?",
			"created_at":      "2026-06-08T17:55:00Z",
		},
	}
	json.NewEncoder(w).Encode(messages)
}

func (h *WebSocketsHandler) SendMessageHTTP(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *WebSocketsHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Notifications
func (h *WebSocketsHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	notes := []map[string]interface{}{
		{
			"id":           "notif_1",
			"recipient_id": "usr_tenant_123",
			"title":        "Viewing Scheduled",
			"body":         "Your viewing for prop_123 was approved for June 15.",
			"read":         false,
			"created_at":   "2026-06-08T17:50:00Z",
		},
	}
	json.NewEncoder(w).Encode(notes)
}

func (h *WebSocketsHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   r.PathValue("id"),
		"read": true,
	})
}

func (h *WebSocketsHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "All notifications marked read."})
}

func (h *WebSocketsHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
