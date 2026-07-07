package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// UsersHandler handles user profile endpoints
type UsersHandler struct {
	db *gorm.DB
}

// NewUsersHandler creates a UsersHandler with DB
func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{db: db}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func userToMap(u models.User) map[string]interface{} {
	return map[string]interface{}{
		"id":             u.ID,
		"first_name":     u.FirstName,
		"last_name":      u.LastName,
		"email":          u.Email,
		"phone":          u.Phone,
		"role":           u.Role,
		"avatar_url":     u.AvatarURL,
		"email_verified": u.EmailVerified,
		"mfa_enabled":    u.MFAEnabled,
		"created_at":     u.CreatedAt,
	}
}

// ─── Me ───────────────────────────────────────────────────────────────────────

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		apierrors.NotFound(w, r, "User not found.")
		return
	}
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	updates := map[string]interface{}{}
	if body.FirstName != "" {
		updates["first_name"] = body.FirstName
	}
	if body.LastName != "" {
		updates["last_name"] = body.LastName
	}
	if body.Phone != "" {
		updates["phone"] = body.Phone
	}

	if len(updates) == 0 {
		apierrors.BadRequest(w, r, "No fields to update.", nil)
		return
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	var user models.User
	h.db.First(&user, "id = ?", userID)
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) DeactivateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if err := h.db.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}
	respond.Message(w, "Account deactivated successfully.")
}

func (h *UsersHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AvatarURL == "" {
		apierrors.BadRequest(w, r, "avatar_url is required.", nil)
		return
	}
	h.db.Model(&models.User{}).Where("id = ?", userID).Update("avatar_url", body.AvatarURL)
	respond.Message(w, "Avatar updated successfully.")
}

// ─── Public Profiles ──────────────────────────────────────────────────────────

func (h *UsersHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var user models.User
	if err := h.db.First(&user, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "User not found.")
		return
	}
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) GetTenantProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", id, "tenant").First(&user).Error; err != nil {
		apierrors.NotFound(w, r, "Tenant profile not found.")
		return
	}
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) UpdateTenantProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	delete(body, "password")
	delete(body, "role")
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(body)
	var user models.User
	h.db.First(&user, "id = ?", userID)
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) GetLandlordProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", id, "landlord").First(&user).Error; err != nil {
		apierrors.NotFound(w, r, "Landlord profile not found.")
		return
	}
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) UpdateLandlordProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	delete(body, "password")
	delete(body, "role")
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(body)
	var user models.User
	h.db.First(&user, "id = ?", userID)
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) GetAgentProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var user models.User
	if err := h.db.Where("id = ? AND role = ?", id, "agent").First(&user).Error; err != nil {
		apierrors.NotFound(w, r, "Agent profile not found.")
		return
	}
	respond.OK(w, userToMap(user))
}

func (h *UsersHandler) UpdateAgentProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	delete(body, "password")
	delete(body, "role")
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(body)
	var user models.User
	h.db.First(&user, "id = ?", userID)
	respond.OK(w, userToMap(user))
}

// ─── Preferences ──────────────────────────────────────────────────────────────

func (h *UsersHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		apierrors.NotFound(w, r, "User not found.")
		return
	}

	var prefs interface{}
	if user.Preferences != "" {
		json.Unmarshal([]byte(user.Preferences), &prefs)
	} else {
		prefs = map[string]interface{}{}
	}
	respond.OK(w, prefs)
}

func (h *UsersHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	raw, _ := json.Marshal(body)
	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("preferences", string(raw)).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}
	respond.Message(w, "Preferences updated successfully.")
}
