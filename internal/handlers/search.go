package handlers

import (
	"encoding/json"
	"net/http"
)

type SearchHandler struct{}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

func (h *SearchHandler) Reindex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Reindexing task queued"})
}

func (h *SearchHandler) GetIndexStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"engine":            "MeiliSearch",
		"status":            "healthy",
		"documents_indexed": 450,
	})
}

func (h *SearchHandler) ListSynonyms(w http.ResponseWriter, r *http.Request) {
	synonyms := []map[string]interface{}{
		{
			"id":       "syn_1",
			"word":     "flat",
			"synonyms": []string{"apartment", "bedsitter", "studio"},
		},
	}
	json.NewEncoder(w).Encode(synonyms)
}

func (h *SearchHandler) CreateSynonym(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *SearchHandler) DeleteSynonym(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Comparison
func (h *SearchHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	compared := []map[string]interface{}{
		{
			"id":    "prop_123456",
			"title": "3 Bedroom House in Ibex Hill",
			"price": 4500,
		},
	}
	json.NewEncoder(w).Encode(compared)
}

func (h *SearchHandler) AddToComparison(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Property added to comparison."})
}

func (h *SearchHandler) RemoveFromComparison(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *SearchHandler) GetComparisonResults(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"properties": []map[string]interface{}{
			{
				"id":    "prop_1",
				"title": "House in Ibex Hill",
				"price": 4500,
				"bedrooms": 3,
			},
			{
				"id":    "prop_2",
				"title": "Flat in Woodlands",
				"price": 5500,
				"bedrooms": 2,
			},
		},
		"differences": []string{"price", "bedrooms", "area"},
	})
}
