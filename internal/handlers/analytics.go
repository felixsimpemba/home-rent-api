package handlers

import (
	"encoding/json"
	"net/http"
)

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

func (h *AnalyticsHandler) GetPropertyAnalytics(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"views":           420,
		"inquiries_count": 18,
		"viewings_count":  5,
	})
}

func (h *AnalyticsHandler) GetLandlordDashboard(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_listings": 3,
		"total_income":    13500.00,
		"occupancy_rate":  0.66,
	})
}

func (h *AnalyticsHandler) GetPlatformOverview(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_listings": 450,
		"active_users":   1200,
		"rented_rate":    0.72,
	})
}

func (h *AnalyticsHandler) GetPlatformRevenue(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"monthly_revenue":     42000.00,
		"total_subscriptions": 120,
	})
}

func (h *AnalyticsHandler) GetPricingTrends(w http.ResponseWriter, r *http.Request) {
	trends := []map[string]interface{}{
		{
			"region":        "Ibex Hill, Lusaka",
			"average_price": 4650.00,
			"year_month":    "2026-05",
		},
		{
			"region":        "Roma, Lusaka",
			"average_price": 9500.00,
			"year_month":    "2026-05",
		},
	}
	json.NewEncoder(w).Encode(trends)
}
