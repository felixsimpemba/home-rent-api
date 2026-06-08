package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/felixsimpemba/home-rent-api/internal/errors"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Password    string `json:"password"`
		AccountType string `json:"account_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	if body.Email == "" || body.Phone == "" || body.Password == "" {
		errors.BadRequest(w, r, "Email, Phone, and Password are required.", []errors.InvalidParam{
			{Name: "email", Reason: "must not be blank"},
			{Name: "phone", Reason: "must not be blank"},
			{Name: "password", Reason: "must not be blank"},
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Registration successful",
		"data": map[string]interface{}{
			"user_id":               "usr_mock_12345",
			"verification_required": true,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.BadRequest(w, r, "Invalid request payload", nil)
		return
	}

	if body.Email == "" || body.Password == "" {
		errors.BadRequest(w, r, "Missing email or password", nil)
		return
	}

	// Mock response
	role := "tenant"
	if body.Email == "admin@homerent.zm" {
		role = "admin"
	} else if body.Email == "agent@homerent.zm" {
		role = "agent"
	} else if body.Email == "landlord@homerent.zm" {
		role = "landlord"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"token":         role + "_token",
		"refresh_token": "refresh_mock_token_abc123",
		"user": map[string]interface{}{
			"id":         "usr_mock_12345",
			"first_name": "Felix",
			"last_name":  "Simpemba",
			"email":      body.Email,
			"role":       role,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Successfully logged out"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"token":         "new_tenant_token",
		"refresh_token": "new_refresh_mock_token_abc123",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Reset link dispatched to email."})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Password updated successfully."})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Email address validated."})
}

func (h *AuthHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"secret":       "NBSWY3DPEB3W64TBNQ",
		"qr_code_url":  "otpauth://totp/HomeRent:felix@example.com?secret=NBSWY3DPEB3W64TBNQ&issuer=HomeRent",
	})
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "MFA disabled"})
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   "tenant_token",
	})
}

func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []map[string]interface{}{
		{
			"id":          "sess_1",
			"device_name": "Chrome on Linux (Ubuntu)",
			"ip_address":  "192.168.1.100",
			"last_active": "2026-06-08T17:55:00Z",
		},
		{
			"id":          "sess_2",
			"device_name": "Safari on iPhone 15 Pro",
			"ip_address":  "10.0.0.5",
			"last_active": "2026-06-08T16:12:00Z",
		},
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Device session terminated"})
}
