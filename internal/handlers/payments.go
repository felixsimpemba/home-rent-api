package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/felixsimpemba/home-rent-api/internal/errors"
)

type PaymentsHandler struct{}

func NewPaymentsHandler() *PaymentsHandler {
	return &PaymentsHandler{}
}

func (h *PaymentsHandler) InitiateMobileMoney(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider    string  `json:"provider"`
		PhoneNumber string  `json:"phone_number"`
		Amount      float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.BadRequest(w, r, "Invalid request body", nil)
		return
	}

	// Mock initiation
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           "tx_mock_momo_987",
		"reference":    "REF-ZAM-MOMO-88",
		"provider":     body.Provider,
		"phone_number": body.PhoneNumber,
		"amount":       body.Amount,
		"currency":     "ZMW",
		"status":       "pending",
	})
}

func (h *PaymentsHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	txs := []map[string]interface{}{
		{
			"id":           "tx_1",
			"reference":    "REF-123",
			"provider":     "mtn",
			"phone_number": "+260977123456",
			"amount":       250.0,
			"currency":     "ZMW",
			"status":       "successful",
			"created_at":   "2026-06-08T15:30:00Z",
		},
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": txs,
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 20,
			"total": 1,
		},
	})
}

func (h *PaymentsHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"reference":    "REF-123",
		"provider":     "mtn",
		"phone_number": "+260977123456",
		"amount":       250.0,
		"currency":     "ZMW",
		"status":       "successful",
	})
}

// Webhooks
func (h *PaymentsHandler) MtnWebhook(w http.ResponseWriter, r *http.Request) {
	// Process MTN callback signature & payload
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "MTN Webhook processed"})
}

func (h *PaymentsHandler) AirtelWebhook(w http.ResponseWriter, r *http.Request) {
	// Process Airtel callback signature & payload
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Airtel Webhook processed"})
}

func (h *PaymentsHandler) ZamtelWebhook(w http.ResponseWriter, r *http.Request) {
	// Process Zamtel callback signature & payload
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Zamtel Webhook processed"})
}

// Refunds
func (h *PaymentsHandler) ListRefunds(w http.ResponseWriter, r *http.Request) {
	refunds := []map[string]interface{}{
		{
			"id":             "ref_1",
			"transaction_id": "tx_1",
			"amount":         250.0,
			"status":         "processed",
		},
	}
	json.NewEncoder(w).Encode(refunds)
}

func (h *PaymentsHandler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *PaymentsHandler) GetRefund(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             id,
		"transaction_id": "tx_1",
		"amount":         250.0,
		"status":         "processed",
	})
}

// Subscriptions
func (h *PaymentsHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	plans := []map[string]interface{}{
		{
			"id":       "plan_agent_basic",
			"name":     "Agent Standard",
			"price":    150.0,
			"features": []string{"Up to 10 active listings", "Standard support"},
		},
		{
			"id":       "plan_agent_premium",
			"name":     "Agent Premium",
			"price":    350.0,
			"features": []string{"Unlimited listings", "CRM Integration", "Priority support"},
		},
	}
	json.NewEncoder(w).Encode(plans)
}

func (h *PaymentsHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "sub_9921",
		"user_id": "usr_agent_789",
		"plan_id": "plan_agent_premium",
		"status":  "active",
		"ends_at": "2026-07-08T00:00:00Z",
	})
}

func (h *PaymentsHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      "sub_mock_123",
		"status":  "active",
		"details": body,
	})
}

func (h *PaymentsHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Auto-renewal cancelled."})
}
