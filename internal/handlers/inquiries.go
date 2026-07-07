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

// InquiriesHandler handles inquiries, viewings, bookings, leases and screenings
type InquiriesHandler struct {
	db *gorm.DB
}

// NewInquiriesHandler creates an InquiriesHandler with DB
func NewInquiriesHandler(db *gorm.DB) *InquiriesHandler {
	return &InquiriesHandler{db: db}
}

// ─── Inquiries ────────────────────────────────────────────────────────────────

func (h *InquiriesHandler) ListInquiries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Inquiry{}).Preload("Property").Preload("Tenant").Preload("Replies")
	if role == "tenant" {
		tx = tx.Where("tenant_id = ?", userID)
	} else if role == "landlord" || role == "agent" {
		tx = tx.Joins("JOIN properties ON properties.id = inquiries.property_id AND properties.owner_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var inquiries []models.Inquiry
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&inquiries)
	respond.Paginated(w, inquiries, page, limit, total)
}

func (h *InquiriesHandler) CreateInquiry(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetUserID(r)
	var body struct {
		PropertyID string `json:"property_id"`
		Message    string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PropertyID == "" || body.Message == "" {
		apierrors.BadRequest(w, r, "property_id and message are required.", nil)
		return
	}
	inquiry := models.Inquiry{
		ID:         "inq_" + uuid.NewString(),
		PropertyID: body.PropertyID,
		TenantID:   tenantID,
		Message:    body.Message,
		Status:     "open",
	}
	h.db.Create(&inquiry)
	h.db.Preload("Property").Preload("Tenant").First(&inquiry, "id = ?", inquiry.ID)
	respond.Created201(w, inquiry)
}

func (h *InquiriesHandler) GetInquiry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var inquiry models.Inquiry
	if err := h.db.Preload("Property").Preload("Tenant").Preload("Replies.Sender").
		First(&inquiry, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Inquiry not found.")
		return
	}
	respond.OK(w, inquiry)
}

func (h *InquiriesHandler) CloseInquiry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.db.Model(&models.Inquiry{}).Where("id = ?", id).Update("status", "closed")
	respond.Message(w, "Inquiry closed.")
}

func (h *InquiriesHandler) ReplyInquiry(w http.ResponseWriter, r *http.Request) {
	senderID := middleware.GetUserID(r)
	id := r.PathValue("id")
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
		apierrors.BadRequest(w, r, "message is required.", nil)
		return
	}
	reply := models.InquiryReply{
		ID:        "rep_" + uuid.NewString(),
		InquiryID: id,
		SenderID:  senderID,
		Message:   body.Message,
	}
	h.db.Create(&reply)
	h.db.Preload("Sender").First(&reply, "id = ?", reply.ID)
	respond.Created201(w, reply)
}

// ─── Viewings ─────────────────────────────────────────────────────────────────

func (h *InquiriesHandler) ListViewings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Viewing{}).Preload("Property").Preload("Tenant")
	if role == "tenant" {
		tx = tx.Where("tenant_id = ?", userID)
	} else if role == "landlord" || role == "agent" {
		tx = tx.Joins("JOIN properties ON properties.id = viewings.property_id AND properties.owner_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var viewings []models.Viewing
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&viewings)
	respond.Paginated(w, viewings, page, limit, total)
}

func (h *InquiriesHandler) CreateViewing(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetUserID(r)
	var body struct {
		PropertyID string `json:"property_id"`
		Date       string `json:"date"`
		StartTime  string `json:"start_time"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PropertyID == "" || body.Date == "" {
		apierrors.BadRequest(w, r, "property_id and date are required.", nil)
		return
	}
	viewing := models.Viewing{
		ID:         "view_" + uuid.NewString(),
		PropertyID: body.PropertyID,
		TenantID:   tenantID,
		Date:       body.Date,
		StartTime:  body.StartTime,
		Notes:      body.Notes,
		Status:     "pending",
	}
	h.db.Create(&viewing)
	h.db.Preload("Property").Preload("Tenant").First(&viewing, "id = ?", viewing.ID)
	respond.Created201(w, viewing)
}

func (h *InquiriesHandler) GetViewing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var viewing models.Viewing
	if err := h.db.Preload("Property").Preload("Tenant").First(&viewing, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Viewing not found.")
		return
	}
	respond.OK(w, viewing)
}

func (h *InquiriesHandler) ApproveViewing(w http.ResponseWriter, r *http.Request) {
	h.db.Model(&models.Viewing{}).Where("id = ?", r.PathValue("id")).Update("status", "approved")
	respond.Message(w, "Viewing approved.")
}

func (h *InquiriesHandler) RejectViewing(w http.ResponseWriter, r *http.Request) {
	h.db.Model(&models.Viewing{}).Where("id = ?", r.PathValue("id")).Update("status", "rejected")
	respond.Message(w, "Viewing rejected.")
}

func (h *InquiriesHandler) RescheduleViewing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Date      string `json:"date"`
		StartTime string `json:"start_time"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	h.db.Model(&models.Viewing{}).Where("id = ?", id).Updates(map[string]interface{}{
		"date":       body.Date,
		"start_time": body.StartTime,
		"status":     "pending",
	})
	respond.Message(w, "Viewing rescheduled.")
}

func (h *InquiriesHandler) SubmitViewingFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Feedback string `json:"feedback"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	h.db.Model(&models.Viewing{}).Where("id = ?", id).Update("feedback", body.Feedback)
	respond.Message(w, "Feedback submitted.")
}

// ─── Bookings ─────────────────────────────────────────────────────────────────

func (h *InquiriesHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Booking{}).Preload("Property").Preload("Tenant")
	if role == "tenant" {
		tx = tx.Where("tenant_id = ?", userID)
	} else if role == "landlord" || role == "agent" {
		tx = tx.Joins("JOIN properties ON properties.id = bookings.property_id AND properties.owner_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var bookings []models.Booking
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&bookings)
	respond.Paginated(w, bookings, page, limit, total)
}

func (h *InquiriesHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetUserID(r)
	var body struct {
		PropertyID string `json:"property_id"`
		MoveInDate string `json:"move_in_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PropertyID == "" {
		apierrors.BadRequest(w, r, "property_id is required.", nil)
		return
	}
	booking := models.Booking{
		ID:         "book_" + uuid.NewString(),
		PropertyID: body.PropertyID,
		TenantID:   tenantID,
		MoveInDate: body.MoveInDate,
		Status:     "pending",
	}
	h.db.Create(&booking)
	h.db.Preload("Property").Preload("Tenant").First(&booking, "id = ?", booking.ID)
	respond.Created201(w, booking)
}

func (h *InquiriesHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var booking models.Booking
	if err := h.db.Preload("Property").Preload("Tenant").First(&booking, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Booking not found.")
		return
	}
	respond.OK(w, booking)
}

func (h *InquiriesHandler) ApproveBooking(w http.ResponseWriter, r *http.Request) {
	h.db.Model(&models.Booking{}).Where("id = ?", r.PathValue("id")).Update("status", "approved")
	respond.Message(w, "Booking approved.")
}

func (h *InquiriesHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	h.db.Model(&models.Booking{}).Where("id = ?", r.PathValue("id")).Update("status", "cancelled")
	respond.Message(w, "Booking cancelled.")
}

// ─── Leases ───────────────────────────────────────────────────────────────────

func (h *InquiriesHandler) ListLeases(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Lease{}).Preload("Property").Preload("Tenant")
	if role == "tenant" {
		tx = tx.Where("tenant_id = ?", userID)
	} else if role == "landlord" || role == "agent" {
		tx = tx.Where("landlord_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var leases []models.Lease
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&leases)
	respond.Paginated(w, leases, page, limit, total)
}

func (h *InquiriesHandler) CreateLease(w http.ResponseWriter, r *http.Request) {
	landlordID := middleware.GetUserID(r)
	var body struct {
		BookingID   string  `json:"booking_id"`
		PropertyID  string  `json:"property_id"`
		TenantID    string  `json:"tenant_id"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		MonthlyRent float64 `json:"monthly_rent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}
	lease := models.Lease{
		ID:          "lease_" + uuid.NewString(),
		BookingID:   body.BookingID,
		PropertyID:  body.PropertyID,
		TenantID:    body.TenantID,
		LandlordID:  landlordID,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
		MonthlyRent: body.MonthlyRent,
		Status:      "draft",
	}
	h.db.Create(&lease)
	respond.Created201(w, lease)
}

func (h *InquiriesHandler) GetLease(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var lease models.Lease
	if err := h.db.Preload("Property").Preload("Tenant").First(&lease, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Lease not found.")
		return
	}
	respond.OK(w, lease)
}

func (h *InquiriesHandler) SignLease(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	id := r.PathValue("id")

	var lease models.Lease
	if err := h.db.First(&lease, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Lease not found.")
		return
	}

	if role == "tenant" {
		h.db.Model(&lease).Update("tenant_signed", true)
	} else {
		h.db.Model(&lease).Update("landlord_signed", true)
	}

	// If both signed, activate
	h.db.First(&lease, "id = ?", id)
	if lease.TenantSigned && lease.LandlordSigned {
		h.db.Model(&lease).Update("status", "signed")
	}
	_ = userID
	respond.Message(w, "Lease signed successfully.")
}

func (h *InquiriesHandler) TerminateLease(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	now := time.Now()
	h.db.Model(&models.Lease{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "terminated",
		"terminated_at": &now,
	})
	respond.Message(w, "Lease terminated.")
}

// ─── Screening ────────────────────────────────────────────────────────────────

func (h *InquiriesHandler) ListScreenings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.TenantScreening{}).Preload("Tenant")
	if role == "tenant" {
		tx = tx.Where("tenant_id = ?", userID)
	} else {
		tx = tx.Where("created_by = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var screenings []models.TenantScreening
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&screenings)
	respond.Paginated(w, screenings, page, limit, total)
}

func (h *InquiriesHandler) GetScreening(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var screening models.TenantScreening
	if err := h.db.Preload("Tenant").First(&screening, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Screening not found.")
		return
	}
	respond.OK(w, screening)
}

func (h *InquiriesHandler) CreateScreening(w http.ResponseWriter, r *http.Request) {
	createdBy := middleware.GetUserID(r)
	var body struct {
		TenantID   string `json:"tenant_id"`
		PropertyID string `json:"property_id"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TenantID == "" {
		apierrors.BadRequest(w, r, "tenant_id is required.", nil)
		return
	}
	screening := models.TenantScreening{
		ID:         "scr_" + uuid.NewString(),
		TenantID:   body.TenantID,
		PropertyID: body.PropertyID,
		Notes:      body.Notes,
		Status:     "pending",
		CreatedBy:  createdBy,
	}
	h.db.Create(&screening)
	respond.Created201(w, screening)
}
