package handlers

import (
	"encoding/json"
	"net/http"
)

type CRMHandler struct{}

func NewCRMHandler() *CRMHandler {
	return &CRMHandler{}
}

// Leads
func (h *CRMHandler) ListLeads(w http.ResponseWriter, r *http.Request) {
	leads := []map[string]interface{}{
		{
			"id":              "lead_1",
			"name":            "John Smith",
			"email":           "johnsmith@example.com",
			"phone":           "+260977222333",
			"pipeline_status": "viewing_scheduled",
			"assigned_properties": []string{"prop_123456"},
		},
	}
	json.NewEncoder(w).Encode(leads)
}

func (h *CRMHandler) CreateLead(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *CRMHandler) GetLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              id,
		"name":            "John Smith",
		"email":           "johnsmith@example.com",
		"pipeline_status": "viewing_scheduled",
	})
}

func (h *CRMHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(body)
}

func (h *CRMHandler) AddLeadNote(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Note added to lead."})
}

// Tasks
func (h *CRMHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := []map[string]interface{}{
		{
			"id":       "task_1",
			"title":    "Follow up on Ibex Hill contract",
			"due_date": "2026-06-12T12:00:00Z",
			"done":     false,
		},
	}
	json.NewEncoder(w).Encode(tasks)
}

func (h *CRMHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *CRMHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	json.NewEncoder(w).Encode(body)
}

func (h *CRMHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Deals
func (h *CRMHandler) GetDeals(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_deals":       5,
		"deal_value_zmw":    25000.0,
	})
}

// Meetings
func (h *CRMHandler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	meetings := []map[string]interface{}{
		{
			"id":           "meet_1",
			"title":        "Contract signing - Woodlands Apartment",
			"meeting_time": "2026-06-15T15:00:00Z",
			"lead_id":      "lead_1",
		},
	}
	json.NewEncoder(w).Encode(meetings)
}

func (h *CRMHandler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

// Performance
func (h *CRMHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leads_converted":  12,
		"conversion_rate":  0.48,
		"avg_days_to_rent": 8.5,
	})
}
