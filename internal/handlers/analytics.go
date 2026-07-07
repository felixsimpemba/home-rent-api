package handlers

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// AnalyticsHandler handles analytics and dashboard endpoints
type AnalyticsHandler struct {
	db *gorm.DB
}

// NewAnalyticsHandler creates an AnalyticsHandler with DB
func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) GetPricingTrends(w http.ResponseWriter, r *http.Request) {
	province := r.URL.Query().Get("province")
	propertyType := r.URL.Query().Get("property_type")

	type PriceTrend struct {
		Province     string  `json:"province"`
		PropertyType string  `json:"property_type"`
		AvgPrice     float64 `json:"avg_price"`
		MinPrice     float64 `json:"min_price"`
		MaxPrice     float64 `json:"max_price"`
		Count        int64   `json:"count"`
	}

	tx := h.db.Model(&models.Property{}).Select(
		"province, property_type, AVG(price) as avg_price, MIN(price) as min_price, MAX(price) as max_price, COUNT(*) as count",
	).Where("status = ?", "active").Group("province, property_type")

	if province != "" {
		tx = tx.Where("province = ?", province)
	}
	if propertyType != "" {
		tx = tx.Where("property_type = ?", propertyType)
	}

	var trends []PriceTrend
	tx.Scan(&trends)
	respond.OK(w, trends)
}

func (h *AnalyticsHandler) GetPropertyAnalytics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var views int64
	var inquiries int64
	var bookings int64
	h.db.Model(&models.Inquiry{}).Where("property_id = ?", id).Count(&inquiries)
	h.db.Model(&models.Booking{}).Where("property_id = ?", id).Count(&bookings)
	h.db.Model(&models.Viewing{}).Where("property_id = ?", id).Count(&views)

	respond.OK(w, map[string]interface{}{
		"property_id": id,
		"views":       views,
		"inquiries":   inquiries,
		"bookings":    bookings,
	})
}

func (h *AnalyticsHandler) GetLandlordDashboard(w http.ResponseWriter, r *http.Request) {
	landlordID := middleware.GetUserID(r)

	var totalListings int64
	var activeListings int64
	var pendingBookings int64
	var totalInquiries int64

	h.db.Model(&models.Property{}).Where("owner_id = ?", landlordID).Count(&totalListings)
	h.db.Model(&models.Property{}).Where("owner_id = ? AND status = ?", landlordID, "active").Count(&activeListings)
	h.db.Model(&models.Booking{}).
		Joins("JOIN properties ON properties.id = bookings.property_id").
		Where("properties.owner_id = ? AND bookings.status = ?", landlordID, "pending").Count(&pendingBookings)
	h.db.Model(&models.Inquiry{}).
		Joins("JOIN properties ON properties.id = inquiries.property_id").
		Where("properties.owner_id = ?", landlordID).Count(&totalInquiries)

	respond.OK(w, map[string]interface{}{
		"total_listings":   totalListings,
		"active_listings":  activeListings,
		"pending_bookings": pendingBookings,
		"total_inquiries":  totalInquiries,
	})
}

func (h *AnalyticsHandler) GetPlatformOverview(w http.ResponseWriter, r *http.Request) {
	var totalUsers int64
	var totalProperties int64
	var totalBookings int64
	var totalTransactions int64

	h.db.Model(&models.User{}).Count(&totalUsers)
	h.db.Model(&models.Property{}).Count(&totalProperties)
	h.db.Model(&models.Booking{}).Count(&totalBookings)
	h.db.Model(&models.Transaction{}).Count(&totalTransactions)

	respond.OK(w, map[string]interface{}{
		"total_users":        totalUsers,
		"total_properties":   totalProperties,
		"total_bookings":     totalBookings,
		"total_transactions": totalTransactions,
	})
}

func (h *AnalyticsHandler) GetPlatformRevenue(w http.ResponseWriter, r *http.Request) {
	type RevenueRow struct {
		Provider string  `json:"provider"`
		Total    float64 `json:"total"`
		Count    int64   `json:"count"`
	}
	var rows []RevenueRow
	h.db.Model(&models.Transaction{}).
		Select("provider, SUM(amount) as total, COUNT(*) as count").
		Where("status = ?", "successful").
		Group("provider").Scan(&rows)

	respond.OK(w, map[string]interface{}{
		"by_provider": rows,
	})
}
