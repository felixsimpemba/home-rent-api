package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// CRMHandler handles agent CRM endpoints
type CRMHandler struct {
	db *gorm.DB
}

// NewCRMHandler creates a CRMHandler with DB
func NewCRMHandler(db *gorm.DB) *CRMHandler {
	return &CRMHandler{db: db}
}

// ─── Leads ────────────────────────────────────────────────────────────────────

func (h *CRMHandler) ListLeads(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Lead{}).Where("agent_id = ?", agentID).Count(&total)

	var leads []models.Lead
	h.db.Where("agent_id = ?", agentID).
		Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&leads)

	respond.Paginated(w, leads, page, limit, total)
}

func (h *CRMHandler) CreateLead(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	var body struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Phone  string `json:"phone"`
		Source string `json:"source"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		apierrors.BadRequest(w, r, "name is required.", nil)
		return
	}

	lead := models.Lead{
		ID:      "lead_" + uuid.NewString(),
		AgentID: agentID,
		Name:    body.Name,
		Email:   body.Email,
		Phone:   body.Phone,
		Source:  body.Source,
		Notes:   body.Notes,
		Status:  "new",
	}
	h.db.Create(&lead)
	respond.Created201(w, lead)
}

func (h *CRMHandler) GetLead(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	id := r.PathValue("id")

	var lead models.Lead
	if err := h.db.Where("id = ? AND agent_id = ?", id, agentID).First(&lead).Error; err != nil {
		apierrors.NotFound(w, r, "Lead not found.")
		return
	}
	respond.OK(w, lead)
}

func (h *CRMHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	id := r.PathValue("id")

	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	delete(body, "agent_id")

	h.db.Model(&models.Lead{}).Where("id = ? AND agent_id = ?", id, agentID).Updates(body)
	var lead models.Lead
	h.db.First(&lead, "id = ?", id)
	respond.OK(w, lead)
}

func (h *CRMHandler) AddLeadNote(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	leadID := r.PathValue("id")

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		apierrors.BadRequest(w, r, "content is required.", nil)
		return
	}

	note := models.LeadNote{
		ID:        "note_" + uuid.NewString(),
		LeadID:    leadID,
		AgentID:   agentID,
		Content:   body.Content,
	}
	h.db.Create(&note)
	respond.Created201(w, note)
}

// ─── Tasks ────────────────────────────────────────────────────────────────────

func (h *CRMHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Task{}).Where("agent_id = ?", agentID).Count(&total)

	var tasks []models.Task
	h.db.Where("agent_id = ?", agentID).
		Offset((page - 1) * limit).Limit(limit).Order("due_date ASC").Find(&tasks)

	respond.Paginated(w, tasks, page, limit, total)
}

func (h *CRMHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		DueDate     *string `json:"due_date"`
		Priority    string  `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		apierrors.BadRequest(w, r, "title is required.", nil)
		return
	}

	task := models.Task{
		ID:          "task_" + uuid.NewString(),
		AgentID:     agentID,
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
	}
	if body.Priority == "" {
		task.Priority = "medium"
	}
	if body.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *body.DueDate)
		if err == nil {
			task.DueDate = &parsed
		}
	}
	h.db.Create(&task)
	respond.Created201(w, task)
}

func (h *CRMHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	id := r.PathValue("id")

	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	delete(body, "agent_id")

	h.db.Model(&models.Task{}).Where("id = ? AND agent_id = ?", id, agentID).Updates(body)
	var task models.Task
	h.db.First(&task, "id = ?", id)
	respond.OK(w, task)
}

func (h *CRMHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Where("id = ? AND agent_id = ?", id, agentID).Delete(&models.Task{})
	respond.NoContent(w)
}

// ─── Deals ────────────────────────────────────────────────────────────────────

func (h *CRMHandler) GetDeals(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)

	var openBookings int64
	var convertedLeads int64
	h.db.Model(&models.Booking{}).
		Joins("JOIN properties ON properties.id = bookings.property_id").
		Where("properties.owner_id = ? AND bookings.status = ?", agentID, "approved").
		Count(&openBookings)
	h.db.Model(&models.Lead{}).Where("agent_id = ? AND status = ?", agentID, "converted").Count(&convertedLeads)

	respond.OK(w, map[string]interface{}{
		"open_deals":      openBookings,
		"converted_leads": convertedLeads,
	})
}

// ─── Meetings ─────────────────────────────────────────────────────────────────

func (h *CRMHandler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	var meetings []models.Meeting
	h.db.Where("agent_id = ?", agentID).Order("start_at ASC").Find(&meetings)
	respond.OK(w, meetings)
}

func (h *CRMHandler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		apierrors.BadRequest(w, r, "title is required.", nil)
		return
	}

	meeting := models.Meeting{
		ID:          "meet_" + uuid.NewString(),
		AgentID:     agentID,
		Title:       body.Title,
		Description: body.Description,
		Location:    body.Location,
	}
	if t, err := time.Parse(time.RFC3339, body.StartAt); err == nil {
		meeting.StartAt = t
	}
	if t, err := time.Parse(time.RFC3339, body.EndAt); err == nil {
		meeting.EndAt = t
	}
	h.db.Create(&meeting)
	respond.Created201(w, meeting)
}

// ─── Performance ──────────────────────────────────────────────────────────────

func (h *CRMHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	agentID := middleware.GetUserID(r)

	var totalLeads int64
	var newLeads int64
	var totalMeetings int64
	var totalListings int64

	h.db.Model(&models.Lead{}).Where("agent_id = ?", agentID).Count(&totalLeads)
	h.db.Model(&models.Lead{}).Where("agent_id = ? AND status = ?", agentID, "new").Count(&newLeads)
	h.db.Model(&models.Meeting{}).Where("agent_id = ?", agentID).Count(&totalMeetings)
	h.db.Model(&models.Property{}).Where("owner_id = ?", agentID).Count(&totalListings)

	respond.OK(w, map[string]interface{}{
		"total_leads":    totalLeads,
		"new_leads":      newLeads,
		"total_meetings": totalMeetings,
		"total_listings": totalListings,
	})
}
