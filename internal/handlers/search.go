package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// SearchHandler handles search index and property comparison endpoints
type SearchHandler struct {
	db *gorm.DB
}

// NewSearchHandler creates a SearchHandler with DB
func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// ─── Property Comparison ──────────────────────────────────────────────────────

// In-memory comparison list per session (for simplicity; in production use Redis)
var comparisonList = []string{}

func (h *SearchHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	if len(comparisonList) == 0 {
		respond.OK(w, []interface{}{})
		return
	}
	var props []models.Property
	h.db.Preload("Images").Preload("Amenities").
		Where("id IN ?", comparisonList).Find(&props)
	respond.OK(w, props)
}

func (h *SearchHandler) AddToComparison(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PropertyID string `json:"property_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PropertyID == "" {
		apierrors.BadRequest(w, r, "property_id is required.", nil)
		return
	}
	if len(comparisonList) >= 4 {
		apierrors.BadRequest(w, r, "You can compare at most 4 properties at a time.", nil)
		return
	}
	for _, id := range comparisonList {
		if id == body.PropertyID {
			respond.Message(w, "Property already in comparison list.")
			return
		}
	}
	comparisonList = append(comparisonList, body.PropertyID)
	respond.Message(w, "Property added to comparison list.")
}

func (h *SearchHandler) RemoveFromComparison(w http.ResponseWriter, r *http.Request) {
	propertyID := r.PathValue("propertyId")
	updated := comparisonList[:0]
	for _, id := range comparisonList {
		if id != propertyID {
			updated = append(updated, id)
		}
	}
	comparisonList = updated
	respond.NoContent(w)
}

func (h *SearchHandler) GetComparisonResults(w http.ResponseWriter, r *http.Request) {
	if len(comparisonList) < 2 {
		apierrors.BadRequest(w, r, "You need at least 2 properties to compare.", nil)
		return
	}
	var props []models.Property
	h.db.Preload("Images").Preload("Amenities").Where("id IN ?", comparisonList).Find(&props)
	respond.OK(w, props)
}

// ─── Search Index Management ──────────────────────────────────────────────────

func (h *SearchHandler) Reindex(w http.ResponseWriter, r *http.Request) {
	var count int64
	h.db.Model(&models.Property{}).Where("status = ?", "active").Count(&count)
	// TODO: Push to Meilisearch / Elasticsearch / Typesense
	respond.OK(w, map[string]interface{}{
		"message":            "Reindex triggered successfully.",
		"properties_indexed": count,
	})
}

func (h *SearchHandler) GetIndexStatus(w http.ResponseWriter, r *http.Request) {
	var total int64
	var active int64
	h.db.Model(&models.Property{}).Count(&total)
	h.db.Model(&models.Property{}).Where("status = ?", "active").Count(&active)
	respond.OK(w, map[string]interface{}{
		"status":              "healthy",
		"total_properties":    total,
		"indexed_properties":  active,
	})
}

func (h *SearchHandler) ListSynonyms(w http.ResponseWriter, r *http.Request) {
	// TODO: Fetch from search engine
	respond.OK(w, []interface{}{})
}

func (h *SearchHandler) CreateSynonym(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	// TODO: Push to search engine
	respond.Created201(w, body)
}

func (h *SearchHandler) DeleteSynonym(w http.ResponseWriter, r *http.Request) {
	// TODO: Delete from search engine
	respond.NoContent(w)
}
