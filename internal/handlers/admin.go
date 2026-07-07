package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// AdminHandler handles admin panel endpoints
type AdminHandler struct {
	db *gorm.DB
}

// NewAdminHandler creates an AdminHandler with DB
func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	q := r.URL.Query()

	tx := h.db.Model(&models.User{})
	if role := q.Get("role"); role != "" {
		tx = tx.Where("role = ?", role)
	}
	if search := q.Get("search"); search != "" {
		like := "%" + search + "%"
		tx = tx.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	tx.Count(&total)
	var users []models.User
	tx.Select("id, first_name, last_name, email, phone, role, email_verified, banned, created_at").
		Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&users)

	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		result = append(result, userToMap(u))
	}
	respond.Paginated(w, result, page, limit, total)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Role == "" {
		apierrors.BadRequest(w, r, "role is required.", nil)
		return
	}

	validRoles := map[string]bool{"tenant": true, "landlord": true, "agent": true, "admin": true, "moderator": true}
	if !validRoles[body.Role] {
		apierrors.BadRequest(w, r, "Invalid role. Must be one of: tenant, landlord, agent, admin, moderator.", nil)
		return
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", id).Update("role", body.Role).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	logAudit(h.db, actorID, "role_change", "user", id, "role changed to "+body.Role, r.RemoteAddr)
	respond.Message(w, "User role updated to "+body.Role+".")
}

func (h *AdminHandler) UpdateUserBan(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	var body struct {
		Banned bool   `json:"banned"`
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	h.db.Model(&models.User{}).Where("id = ?", id).Update("banned", body.Banned)

	action := "unbanned"
	if body.Banned {
		action = "banned"
	}
	logAudit(h.db, actorID, action, "user", id, body.Reason, r.RemoteAddr)
	respond.Message(w, "User "+action+" successfully.")
}

// ─── Property Moderation ──────────────────────────────────────────────────────

func (h *AdminHandler) ListModerationProperties(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	var total int64
	h.db.Model(&models.Property{}).Where("status = ?", status).Count(&total)

	var props []models.Property
	h.db.Where("status = ?", status).Preload("Owner").
		Offset((page - 1) * limit).Limit(limit).Order("created_at ASC").Find(&props)

	respond.Paginated(w, props, page, limit, total)
}

func (h *AdminHandler) ApproveProperty(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Model(&models.Property{}).Where("id = ?", id).Update("status", "active")
	logAudit(h.db, actorID, "property_approved", "property", id, "", r.RemoteAddr)
	respond.Message(w, "Property approved and made active.")
}

func (h *AdminHandler) RejectProperty(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	h.db.Model(&models.Property{}).Where("id = ?", id).Update("status", "rejected")
	logAudit(h.db, actorID, "property_rejected", "property", id, body.Reason, r.RemoteAddr)
	respond.Message(w, "Property rejected.")
}

// ─── Reports ──────────────────────────────────────────────────────────────────

func (h *AdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}

	var total int64
	h.db.Model(&models.Report{}).Where("status = ?", status).Count(&total)

	var reports []models.Report
	h.db.Where("status = ?", status).
		Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&reports)

	respond.Paginated(w, reports, page, limit, total)
}

func (h *AdminHandler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Model(&models.Report{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_by": actorID,
	})
	respond.Message(w, "Report resolved.")
}

// ─── Verifications ────────────────────────────────────────────────────────────

func (h *AdminHandler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	var total int64
	h.db.Model(&models.User{}).Where("email_verified = false AND banned = false").Count(&total)

	var users []models.User
	h.db.Where("email_verified = false AND banned = false").
		Select("id, first_name, last_name, email, role, created_at").
		Offset((page - 1) * limit).Limit(limit).Find(&users)

	respond.Paginated(w, users, page, limit, total)
}

func (h *AdminHandler) ApproveVerification(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Model(&models.User{}).Where("id = ?", id).Update("email_verified", true)
	logAudit(h.db, actorID, "verification_approved", "user", id, "", r.RemoteAddr)
	respond.Message(w, "User verification approved.")
}

func (h *AdminHandler) RejectVerification(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.GetUserID(r)
	id := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	logAudit(h.db, actorID, "verification_rejected", "user", id, body.Reason, r.RemoteAddr)
	respond.Message(w, "User verification rejected.")
}

// ─── Audit Logs ───────────────────────────────────────────────────────────────

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	q := r.URL.Query()

	tx := h.db.Model(&models.AuditLog{})
	if resource := q.Get("resource"); resource != "" {
		tx = tx.Where("resource = ?", resource)
	}
	if actorID := q.Get("actor_id"); actorID != "" {
		tx = tx.Where("actor_id = ?", actorID)
	}

	var total int64
	tx.Count(&total)
	var logs []models.AuditLog
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&logs)
	respond.Paginated(w, logs, page, limit, total)
}

// ─── System Settings ──────────────────────────────────────────────────────────

func (h *AdminHandler) GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	var settings []models.SystemSetting
	h.db.Find(&settings)

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	respond.OK(w, result)
}

func (h *AdminHandler) UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	for key, value := range body {
		h.db.Save(&models.SystemSetting{Key: key, Value: value})
	}
	respond.Message(w, "Settings updated successfully.")
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func logAudit(db *gorm.DB, actorID, action, resource, resourceID, details, ip string) {
	db.Create(&models.AuditLog{
		ID:         "audit_" + uuid.NewString(),
		ActorID:    actorID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ip,
	})
}
