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

// FavoritesHandler handles saved properties and search preferences
type FavoritesHandler struct {
	db *gorm.DB
}

// NewFavoritesHandler creates a FavoritesHandler with DB
func NewFavoritesHandler(db *gorm.DB) *FavoritesHandler {
	return &FavoritesHandler{db: db}
}

// ─── Favorites ────────────────────────────────────────────────────────────────

func (h *FavoritesHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.Favorite{}).Where("user_id = ?", userID).Count(&total)

	var favorites []models.Favorite
	h.db.Where("user_id = ?", userID).
		Offset((page - 1) * limit).Limit(limit).
		Order("created_at DESC").Find(&favorites)

	// Preload property data
	result := make([]map[string]interface{}, 0, len(favorites))
	for _, fav := range favorites {
		var prop models.Property
		h.db.Preload("Images").First(&prop, "id = ?", fav.PropertyID)
		result = append(result, map[string]interface{}{
			"favorite_id":  fav.ID,
			"saved_at":     fav.CreatedAt,
			"property":     prop,
		})
	}

	respond.Paginated(w, result, page, limit, total)
}

func (h *FavoritesHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		PropertyID string `json:"property_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PropertyID == "" {
		apierrors.BadRequest(w, r, "property_id is required.", nil)
		return
	}

	// Check if already favorited
	var existing models.Favorite
	if h.db.Where("user_id = ? AND property_id = ?", userID, body.PropertyID).First(&existing).Error == nil {
		respond.OK(w, map[string]interface{}{"message": "Already in favorites.", "favorite_id": existing.ID})
		return
	}

	fav := models.Favorite{
		ID:         "fav_" + uuid.NewString(),
		UserID:     userID,
		PropertyID: body.PropertyID,
	}
	h.db.Create(&fav)
	respond.Created201(w, fav)
}

func (h *FavoritesHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	propertyID := r.PathValue("propertyId")
	h.db.Where("user_id = ? AND property_id = ?", userID, propertyID).Delete(&models.Favorite{})
	respond.NoContent(w)
}

// ─── Saved Searches ───────────────────────────────────────────────────────────

func (h *FavoritesHandler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var searches []models.SavedSearch
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&searches)
	respond.OK(w, searches)
}

func (h *FavoritesHandler) SaveSearch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		Name    string                 `json:"name"`
		Filters map[string]interface{} `json:"filters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		apierrors.BadRequest(w, r, "name and filters are required.", nil)
		return
	}

	filtersJSON, _ := json.Marshal(body.Filters)
	search := models.SavedSearch{
		ID:      "ss_" + uuid.NewString(),
		UserID:  userID,
		Name:    body.Name,
		Filters: string(filtersJSON),
	}
	h.db.Create(&search)
	respond.Created201(w, search)
}

func (h *FavoritesHandler) DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := r.PathValue("id")
	h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SavedSearch{})
	respond.NoContent(w)
}
