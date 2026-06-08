package handlers

import (
	"encoding/json"
	"net/http"
)

type UsersHandler struct{}

func NewUsersHandler() *UsersHandler {
	return &UsersHandler{}
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         "usr_mock_12345",
		"first_name": "Felix",
		"last_name":  "Simpemba",
		"email":      "felix@example.com",
		"phone":      "+260977123456",
		"role":       "tenant",
	})
}

func (h *UsersHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         "usr_mock_12345",
		"first_name": body["first_name"],
		"last_name":  body["last_name"],
		"email":      "felix@example.com",
		"phone":      body["phone"],
		"role":       "tenant",
	})
}

func (h *UsersHandler) DeactivateMe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"phone":      "+260977111222",
		"role":       "landlord",
	})
}

func (h *UsersHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"avatar_url": "https://cdn.site.com/avatars/new_usr_12345.jpg",
	})
}

func (h *UsersHandler) GetTenantProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      id,
		"occupation":   "Software Developer",
		"nrc_number":   "123456/11/1",
		"references":   []string{"ref_1", "ref_2"},
	})
}

func (h *UsersHandler) UpdateTenantProfile(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *UsersHandler) GetLandlordProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      id,
		"company_name": "Lsk Rentals Ltd",
		"nrc_number":   "987654/10/1",
		"verified":     true,
	})
}

func (h *UsersHandler) UpdateLandlordProfile(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *UsersHandler) GetAgentProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":        id,
		"agency_name":    "Ibex Properties",
		"license_number": "LIC-AGENT-890",
		"tax_number":     "TPIN-98782",
		"verified":       true,
	})
}

func (h *UsersHandler) UpdateAgentProfile(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *UsersHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email_notifications": true,
		"push_notifications":  true,
		"preferred_provinces": []string{"Lusaka", "Copperbelt"},
	})
}

func (h *UsersHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}
