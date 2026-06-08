package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type PropertiesHandler struct{}

func NewPropertiesHandler() *PropertiesHandler {
	return &PropertiesHandler{}
}

func (h *PropertiesHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	properties := []map[string]interface{}{
		{
			"id":             "prop_123456",
			"title":          "3 Bedroom House in Ibex Hill",
			"description":    "Beautiful family house with large garden",
			"property_type":  "house",
			"listing_type":   "rent",
			"price":          4500,
			"deposit":        4500,
			"bedrooms":       3,
			"bathrooms":      2,
			"parking_spaces": 2,
			"province":       "Lusaka",
			"district":       "Lusaka",
			"area":           "Ibex Hill",
			"latitude":       -15.4067,
			"longitude":      28.2871,
			"amenities":      []string{"water", "electricity", "borehole"},
			"status":         "active",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": properties,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": 1,
		},
	})
}

func (h *PropertiesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"property_id": "prop_mock_" + strconv.FormatInt(timeNowUnix(), 10),
		"data":        body,
	})
}

func (h *PropertiesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             id,
		"title":          "3 Bedroom House in Ibex Hill",
		"description":    "Beautiful family house with large garden",
		"property_type":  "house",
		"listing_type":   "rent",
		"price":          4500,
		"deposit":        4500,
		"bedrooms":       3,
		"bathrooms":      2,
		"parking_spaces": 2,
		"province":       "Lusaka",
		"district":       "Lusaka",
		"area":           "Ibex Hill",
		"latitude":       -15.4067,
		"longitude":      28.2871,
		"amenities":      []string{"water", "electricity", "borehole"},
		"status":         "active",
		"owner": map[string]interface{}{
			"id":   "usr_landlord_456",
			"name": "Jane Landlord",
		},
	})
}

func (h *PropertiesHandler) Update(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    body,
	})
}

func (h *PropertiesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *PropertiesHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  body.Status,
	})
}

func (h *PropertiesHandler) UploadImages(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"images": []map[string]interface{}{
			{
				"id":         "img_1",
				"url":        "https://cdn.site.com/properties/img1.jpg",
				"created_at": "2026-06-08T17:55:00Z",
			},
		},
	})
}

func (h *PropertiesHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *PropertiesHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	availability := []map[string]interface{}{
		{
			"id":         "block_1",
			"date":       "2026-06-15",
			"start_time": "10:00",
			"end_time":   "12:00",
		},
	}
	json.NewEncoder(w).Encode(availability)
}

func (h *PropertiesHandler) AddAvailabilityBlock(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *PropertiesHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	featured := []map[string]interface{}{
		{
			"id":    "prop_featured_1",
			"title": "Luxury Apartment in Roma",
			"price": 12000,
		},
	}
	json.NewEncoder(w).Encode(featured)
}

func (h *PropertiesHandler) GetNearby(w http.ResponseWriter, r *http.Request) {
	nearby := []map[string]interface{}{
		{
			"id":    "prop_nearby_1",
			"title": "Cozy Flat in Ibex Hill",
			"price": 3500,
		},
	}
	json.NewEncoder(w).Encode(nearby)
}

func (h *PropertiesHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	recs := []map[string]interface{}{
		{
			"id":    "prop_rec_1",
			"title": "2 Bedroom Apartment close to Woodlands",
			"price": 5500,
		},
	}
	json.NewEncoder(w).Encode(recs)
}

func (h *PropertiesHandler) ReportListing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Report logged."})
}

func (h *PropertiesHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	history := []map[string]interface{}{
		{
			"id":        "hist_1",
			"event":     "price_change",
			"old_value": "4800",
			"new_value": "4500",
		},
	}
	json.NewEncoder(w).Encode(history)
}

func (h *PropertiesHandler) UpdateAmenities(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Amenities updated."})
}

// Categories management
func (h *PropertiesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats := []map[string]interface{}{
		{"id": "cat_1", "name": "Apartment", "slug": "apartment"},
		{"id": "cat_2", "name": "House", "slug": "house"},
		{"id": "cat_3", "name": "Studio", "slug": "studio"},
	}
	json.NewEncoder(w).Encode(cats)
}

func (h *PropertiesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *PropertiesHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *PropertiesHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func timeNowUnix() int64 {
	return 1775700000 // Stable mock timestamp
}
