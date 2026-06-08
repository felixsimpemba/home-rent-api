package handlers

import (
	"encoding/json"
	"net/http"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := []map[string]interface{}{
		{
			"id":         "usr_tenant_123",
			"first_name": "Felix",
			"last_name":  "Simpemba",
			"email":      "felix@example.com",
			"role":       "tenant",
		},
	}
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Role string `json:"role"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      r.PathValue("id"),
		"success": true,
		"role":    body.Role,
	})
}

func (h *AdminHandler) UpdateUserBan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Banned bool   `json:"banned"`
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      r.PathValue("id"),
		"success": true,
		"banned":  body.Banned,
		"reason":  body.Reason,
	})
}

func (h *AdminHandler) ListModerationProperties(w http.ResponseWriter, r *http.Request) {
	props := []map[string]interface{}{
		{
			"id":     "prop_123456",
			"title":  "3 Bedroom House in Ibex Hill",
			"status": "pending",
		},
	}
	json.NewEncoder(w).Encode(props)
}

func (h *AdminHandler) ApproveProperty(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      r.PathValue("id"),
		"status":  "active",
		"success": true,
	})
}

func (h *AdminHandler) RejectProperty(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      r.PathValue("id"),
		"status":  "rejected",
		"reason":  body.Reason,
		"success": true,
	})
}

func (h *AdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	reports := []map[string]interface{}{
		{
			"id":          "rep_1",
			"reporter_id": "usr_tenant_123",
			"target_type": "property",
			"target_id":   "prop_123456",
			"reason":      "inaccurate pricing",
			"status":      "pending",
		},
	}
	json.NewEncoder(w).Encode(reports)
}

func (h *AdminHandler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      r.PathValue("id"),
		"status":  "resolved",
		"success": true,
	})
}

func (h *AdminHandler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	items := []map[string]interface{}{
		{
			"id":            "ver_1",
			"user_id":       "usr_landlord_456",
			"role":          "landlord",
			"document_urls": []string{"https://cdn.site.com/nrc.jpg"},
			"submitted_at":  "2026-06-08T15:00:00Z",
		},
	}
	json.NewEncoder(w).Encode(items)
}

func (h *AdminHandler) ApproveVerification(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "KYC Approved"})
}

func (h *AdminHandler) RejectVerification(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "KYC Rejected"})
}

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs := []map[string]interface{}{
		{
			"id":        "log_1",
			"actor_id":  "usr_admin_000",
			"action":    "approve_verification",
			"details":   "Approved KYC for usr_landlord_456",
			"timestamp": "2026-06-08T17:55:00Z",
		},
	}
	json.NewEncoder(w).Encode(logs)
}

func (h *AdminHandler) GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"maintenance_mode":            false,
		"listing_moderation_required": true,
		"agent_subscription_price":    350.00,
	})
}

func (h *AdminHandler) UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(body)
}
