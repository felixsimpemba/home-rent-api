package handlers

import (
	"encoding/json"
	"net/http"
)

type InquiriesHandler struct{}

func NewInquiriesHandler() *InquiriesHandler {
	return &InquiriesHandler{}
}

// Inquiries
func (h *InquiriesHandler) ListInquiries(w http.ResponseWriter, r *http.Request) {
	inquiries := []map[string]interface{}{
		{
			"id":          "inq_123",
			"property_id": "prop_123",
			"tenant_id":   "usr_tenant_123",
			"message":     "Is this house still available?",
			"status":      "open",
			"created_at":  "2026-06-08T17:55:00Z",
		},
	}
	json.NewEncoder(w).Encode(inquiries)
}

func (h *InquiriesHandler) CreateInquiry(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *InquiriesHandler) GetInquiry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"property_id": "prop_123",
		"message":     "Is this house still available?",
		"status":      "open",
	})
}

func (h *InquiriesHandler) CloseInquiry(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "closed"})
}

func (h *InquiriesHandler) ReplyInquiry(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Reply added."})
}

// Viewings
func (h *InquiriesHandler) ListViewings(w http.ResponseWriter, r *http.Request) {
	viewings := []map[string]interface{}{
		{
			"id":          "view_1",
			"property_id": "prop_123",
			"tenant_id":   "usr_tenant_123",
			"date":        "2026-06-15",
			"time":        "10:00",
			"status":      "pending",
		},
	}
	json.NewEncoder(w).Encode(viewings)
}

func (h *InquiriesHandler) CreateViewing(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *InquiriesHandler) GetViewing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"property_id": "prop_123",
		"date":        "2026-06-15",
		"time":        "10:00",
		"status":      "pending",
	})
}

func (h *InquiriesHandler) ApproveViewing(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "approved"})
}

func (h *InquiriesHandler) RejectViewing(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "rejected"})
}

func (h *InquiriesHandler) RescheduleViewing(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "rescheduled", "new_time": body})
}

func (h *InquiriesHandler) SubmitViewingFeedback(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Feedback submitted."})
}

// Bookings
func (h *InquiriesHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	bookings := []map[string]interface{}{
		{
			"id":          "bk_1",
			"property_id": "prop_123",
			"tenant_id":   "usr_tenant_123",
			"status":      "pending",
		},
	}
	json.NewEncoder(w).Encode(bookings)
}

func (h *InquiriesHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *InquiriesHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"property_id": "prop_123",
		"status":      "pending",
	})
}

func (h *InquiriesHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "cancelled"})
}

func (h *InquiriesHandler) ApproveBooking(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "approved"})
}

// Leases
func (h *InquiriesHandler) ListLeases(w http.ResponseWriter, r *http.Request) {
	leases := []map[string]interface{}{
		{
			"id":              "lease_1",
			"booking_id":      "bk_1",
			"terms":           "Standard rent agreement...",
			"tenant_signed":   false,
			"landlord_signed": false,
		},
	}
	json.NewEncoder(w).Encode(leases)
}

func (h *InquiriesHandler) CreateLease(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *InquiriesHandler) GetLease(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"booking_id": "bk_1",
		"terms":      "Standard rent agreement...",
	})
}

func (h *InquiriesHandler) SignLease(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "signed"})
}

func (h *InquiriesHandler) TerminateLease(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "status": "termination_requested"})
}

// Screenings
func (h *InquiriesHandler) ListScreenings(w http.ResponseWriter, r *http.Request) {
	screenings := []map[string]interface{}{
		{
			"id":        "scr_1",
			"tenant_id": "usr_tenant_123",
			"status":    "completed",
		},
	}
	json.NewEncoder(w).Encode(screenings)
}

func (h *InquiriesHandler) CreateScreening(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *InquiriesHandler) GetScreening(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"tenant_id":  "usr_tenant_123",
		"status":     "completed",
		"report_url": "https://cdn.site.com/screening/rep_1.pdf",
	})
}
