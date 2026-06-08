package handlers

import (
	"encoding/json"
	"net/http"
)

type FavoritesHandler struct{}

func NewFavoritesHandler() *FavoritesHandler {
	return &FavoritesHandler{}
}

func (h *FavoritesHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	favs := []map[string]interface{}{
		{
			"id":    "prop_123456",
			"title": "3 Bedroom House in Ibex Hill",
			"price": 4500,
		},
	}
	json.NewEncoder(w).Encode(favs)
}

func (h *FavoritesHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Property added to favorites."})
}

func (h *FavoritesHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoritesHandler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	searches := []map[string]interface{}{
		{
			"id":    "search_1",
			"title": "Ibex Hill 3 Bed Under 5000",
			"filters": map[string]interface{}{
				"area":      "Ibex Hill",
				"max_price": 5000,
				"bedrooms":  3,
			},
		},
	}
	json.NewEncoder(w).Encode(searches)
}

func (h *FavoritesHandler) SaveSearch(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *FavoritesHandler) DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
