package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// PropertiesHandler handles property listing endpoints
type PropertiesHandler struct {
	db *gorm.DB
}

// NewPropertiesHandler creates a PropertiesHandler with DB
func NewPropertiesHandler(db *gorm.DB) *PropertiesHandler {
	return &PropertiesHandler{db: db}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func pageLimit(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return page, limit
}

// ─── Search / List ────────────────────────────────────────────────────────────

func (h *PropertiesHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	q := r.URL.Query()

	tx := h.db.Model(&models.Property{}).Where("status = ?", "active").
		Preload("Owner").Preload("Images").Preload("Amenities")

	if v := q.Get("province"); v != "" {
		tx = tx.Where("province = ?", v)
	}
	if v := q.Get("district"); v != "" {
		tx = tx.Where("district = ?", v)
	}
	if v := q.Get("area"); v != "" {
		tx = tx.Where("area LIKE ?", "%"+v+"%")
	}
	if v := q.Get("property_type"); v != "" {
		tx = tx.Where("property_type = ?", v)
	}
	if v := q.Get("listing_type"); v != "" {
		tx = tx.Where("listing_type = ?", v)
	}
	if v := q.Get("bedrooms"); v != "" {
		tx = tx.Where("bedrooms >= ?", v)
	}
	if v := q.Get("min_price"); v != "" {
		tx = tx.Where("price >= ?", v)
	}
	if v := q.Get("max_price"); v != "" {
		tx = tx.Where("price <= ?", v)
	}

	var total int64
	tx.Count(&total)

	var properties []models.Property
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&properties)

	respond.Paginated(w, properties, page, limit, total)
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.GetUserID(r)

	var body struct {
		Title         string   `json:"title"`
		Description   string   `json:"description"`
		PropertyType  string   `json:"property_type"`
		ListingType   string   `json:"listing_type"`
		Price         float64  `json:"price"`
		Deposit       float64  `json:"deposit"`
		Bedrooms      int      `json:"bedrooms"`
		Bathrooms     int      `json:"bathrooms"`
		ParkingSpaces int      `json:"parking_spaces"`
		Province      string   `json:"province"`
		District      string   `json:"district"`
		Area          string   `json:"area"`
		Address       string   `json:"address"`
		Latitude      float64  `json:"latitude"`
		Longitude     float64  `json:"longitude"`
		Amenities     []string `json:"amenities"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	if body.Title == "" || body.Province == "" || body.District == "" || body.PropertyType == "" || body.Price <= 0 {
		apierrors.BadRequest(w, r, "title, province, district, property_type, and price are required.", nil)
		return
	}

	prop := models.Property{
		ID:            "prop_" + uuid.NewString(),
		Title:         body.Title,
		Description:   body.Description,
		PropertyType:  body.PropertyType,
		ListingType:   body.ListingType,
		Price:         body.Price,
		Deposit:       body.Deposit,
		Bedrooms:      body.Bedrooms,
		Bathrooms:     body.Bathrooms,
		ParkingSpaces: body.ParkingSpaces,
		Province:      body.Province,
		District:      body.District,
		Area:          body.Area,
		Address:       body.Address,
		Latitude:      body.Latitude,
		Longitude:     body.Longitude,
		Status:        "pending", // requires admin approval
		OwnerID:       ownerID,
	}

	if err := h.db.Create(&prop).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	// Create amenities
	for _, name := range body.Amenities {
		a := models.Amenity{
			ID:         "amen_" + uuid.NewString(),
			PropertyID: prop.ID,
			Name:       name,
		}
		h.db.Create(&a)
	}

	// Log creation in history
	h.db.Create(&models.PropertyHistoryLog{
		ID:         "hist_" + uuid.NewString(),
		PropertyID: prop.ID,
		Event:      "created",
		NewValue:   prop.Title,
		ChangedBy:  ownerID,
	})

	h.db.Preload("Owner").Preload("Amenities").First(&prop, "id = ?", prop.ID)
	respond.Created201(w, prop)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var prop models.Property
	if err := h.db.Preload("Owner").Preload("Images").Preload("Amenities").
		First(&prop, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Property not found.")
		return
	}
	respond.OK(w, prop)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	id := r.PathValue("id")

	var prop models.Property
	if err := h.db.First(&prop, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Property not found.")
		return
	}

	// Ownership check (admins can bypass)
	if role != "admin" && prop.OwnerID != userID {
		apierrors.Forbidden(w, r, "You do not own this property.")
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	// Strip protected fields
	delete(body, "id")
	delete(body, "owner_id")
	delete(body, "status")
	delete(body, "created_at")

	if err := h.db.Model(&prop).Updates(body).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	h.db.Create(&models.PropertyHistoryLog{
		ID:         "hist_" + uuid.NewString(),
		PropertyID: prop.ID,
		Event:      "updated",
		ChangedBy:  userID,
	})

	h.db.Preload("Owner").Preload("Images").Preload("Amenities").First(&prop, "id = ?", id)
	respond.OK(w, prop)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	id := r.PathValue("id")

	var prop models.Property
	if err := h.db.First(&prop, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Property not found.")
		return
	}

	if role != "admin" && prop.OwnerID != userID {
		apierrors.Forbidden(w, r, "You do not own this property.")
		return
	}

	h.db.Delete(&prop)
	respond.NoContent(w)
}

// ─── Status ───────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := r.PathValue("id")

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		apierrors.BadRequest(w, r, "status is required.", nil)
		return
	}

	var prop models.Property
	if err := h.db.First(&prop, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Property not found.")
		return
	}

	oldStatus := prop.Status
	h.db.Model(&prop).Update("status", body.Status)

	h.db.Create(&models.PropertyHistoryLog{
		ID:         "hist_" + uuid.NewString(),
		PropertyID: prop.ID,
		Event:      "status_change",
		OldValue:   oldStatus,
		NewValue:   body.Status,
		ChangedBy:  userID,
	})

	respond.OK(w, map[string]interface{}{"id": id, "status": body.Status})
}

// ─── Images ───────────────────────────────────────────────────────────────────

func (h *PropertiesHandler) UploadImages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Images []struct {
			URL     string `json:"url"`
			Caption string `json:"caption"`
		} `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	saved := make([]models.PropertyImage, 0, len(body.Images))
	for i, img := range body.Images {
		pi := models.PropertyImage{
			ID:         "img_" + uuid.NewString(),
			PropertyID: id,
			URL:        img.URL,
			Caption:    img.Caption,
			SortOrder:  i,
		}
		h.db.Create(&pi)
		saved = append(saved, pi)
	}

	respond.Created201(w, saved)
}

func (h *PropertiesHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("imageId")
	h.db.Delete(&models.PropertyImage{}, "id = ?", imageID)
	respond.NoContent(w)
}

// ─── Availability ─────────────────────────────────────────────────────────────

func (h *PropertiesHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var blocks []models.AvailabilityBlock
	h.db.Where("property_id = ?", id).Find(&blocks)
	respond.OK(w, blocks)
}

func (h *PropertiesHandler) AddAvailabilityBlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Date      string `json:"date"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}
	block := models.AvailabilityBlock{
		ID:         "avail_" + uuid.NewString(),
		PropertyID: id,
		Date:       body.Date,
		StartTime:  body.StartTime,
		EndTime:    body.EndTime,
	}
	h.db.Create(&block)
	respond.Created201(w, block)
}

// ─── Featured / Nearby / Recommendations ─────────────────────────────────────

func (h *PropertiesHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	var props []models.Property
	h.db.Where("featured = true AND status = ?", "active").
		Preload("Images").Limit(10).Order("created_at DESC").Find(&props)
	respond.OK(w, props)
}

func (h *PropertiesHandler) GetNearby(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat := q.Get("lat")
	lng := q.Get("lng")
	radius := q.Get("radius")
	if radius == "" {
		radius = "5"
	}

	var props []models.Property
	if lat != "" && lng != "" {
		// Haversine formula via MySQL
		h.db.Raw(`
			SELECT *, (
				6371 * ACOS(
					COS(RADIANS(?)) * COS(RADIANS(latitude)) *
					COS(RADIANS(longitude) - RADIANS(?)) +
					SIN(RADIANS(?)) * SIN(RADIANS(latitude))
				)
			) AS distance
			FROM properties
			WHERE status = 'active' AND deleted_at IS NULL
			HAVING distance < ?
			ORDER BY distance ASC
			LIMIT 20
		`, lat, lng, lat, radius).Scan(&props)
	} else {
		h.db.Where("status = ?", "active").Limit(20).Find(&props)
	}
	respond.OK(w, props)
}

func (h *PropertiesHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement ML-based recommendations; currently returns featured listings
	var props []models.Property
	h.db.Where("status = ?", "active").Preload("Images").
		Order("RAND()").Limit(10).Find(&props)
	respond.OK(w, props)
}

// ─── Reports / History / Amenities ───────────────────────────────────────────

func (h *PropertiesHandler) ReportListing(w http.ResponseWriter, r *http.Request) {
	reporterID := middleware.GetUserID(r)
	propertyID := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	report := models.Report{
		ID:         "rpt_" + uuid.NewString(),
		PropertyID: propertyID,
		ReporterID: reporterID,
		Reason:     body.Reason,
	}
	h.db.Create(&report)
	respond.Created201(w, map[string]interface{}{"report_id": report.ID, "message": "Report submitted successfully."})
}

func (h *PropertiesHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var logs []models.PropertyHistoryLog
	h.db.Where("property_id = ?", id).Order("created_at DESC").Find(&logs)
	respond.OK(w, logs)
}

func (h *PropertiesHandler) UpdateAmenities(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Amenities []string `json:"amenities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	// Replace all amenities
	h.db.Where("property_id = ?", id).Delete(&models.Amenity{})
	for _, name := range body.Amenities {
		h.db.Create(&models.Amenity{
			ID:         "amen_" + uuid.NewString(),
			PropertyID: id,
			Name:       name,
		})
	}
	respond.Message(w, "Amenities updated successfully.")
}

// ─── Categories ───────────────────────────────────────────────────────────────

func (h *PropertiesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	var cats []models.Category
	h.db.Find(&cats)
	respond.OK(w, cats)
}

func (h *PropertiesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Slug == "" {
		apierrors.BadRequest(w, r, "name and slug are required.", nil)
		return
	}
	cat := models.Category{
		ID:   "cat_" + uuid.NewString(),
		Name: body.Name,
		Slug: body.Slug,
	}
	h.db.Create(&cat)
	respond.Created201(w, cat)
}

func (h *PropertiesHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body)
	delete(body, "id")
	h.db.Model(&models.Category{}).Where("id = ?", id).Updates(body)
	var cat models.Category
	h.db.First(&cat, "id = ?", id)
	respond.OK(w, cat)
}

func (h *PropertiesHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.db.Delete(&models.Category{}, "id = ?", id)
	respond.NoContent(w)
}
