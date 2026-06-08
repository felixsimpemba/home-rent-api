package router

import (
	"net/http"

	"github.com/felixsimpemba/home-rent-api/internal/config"
	"github.com/felixsimpemba/home-rent-api/internal/handlers"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
)

type Middleware func(http.Handler) http.Handler

func chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RegisterRoutes registers all endpoints and chains security middlewares
func RegisterRoutes(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	auth := handlers.NewAuthHandler()
	users := handlers.NewUsersHandler()
	properties := handlers.NewPropertiesHandler()
	favorites := handlers.NewFavoritesHandler()
	inquiries := handlers.NewInquiriesHandler()
	payments := handlers.NewPaymentsHandler()
	ws := handlers.NewWebSocketsHandler()
	crm := handlers.NewCRMHandler()
	admin := handlers.NewAdminHandler()
	analytics := handlers.NewAnalyticsHandler()
	uploads := handlers.NewUploadsHandler()
	search := handlers.NewSearchHandler()

	// Base middlewares
	authMW := middleware.Authenticate(cfg.JWTSecret)
	roleTenant := middleware.RequireRoles("tenant", "admin")
	roleLandlord := middleware.RequireRoles("landlord", "admin")
	roleAgent := middleware.RequireRoles("agent", "admin")
	roleAgentLandlord := middleware.RequireRoles("agent", "landlord", "admin")
	roleAdminMod := middleware.RequireRoles("admin", "moderator")
	roleAdminOnly := middleware.RequireRoles("admin")

	// ==========================================
	// 1. PUBLIC ROUTES
	// ==========================================
	mux.HandleFunc("POST /api/v1/auth/register", auth.Register)
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", auth.Refresh)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", auth.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", auth.ResetPassword)
	mux.HandleFunc("POST /api/v1/auth/verify-email", auth.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/mfa/verify", auth.VerifyMFA)

	mux.HandleFunc("GET /api/v1/properties", properties.Search)
	mux.HandleFunc("GET /api/v1/properties/featured", properties.GetFeatured)
	mux.HandleFunc("GET /api/v1/properties/nearby", properties.GetNearby)
	mux.HandleFunc("GET /api/v1/properties/{id}", properties.Get)
	mux.HandleFunc("GET /api/v1/properties/{id}/availability", properties.GetAvailability)
	mux.HandleFunc("GET /api/v1/properties/categories", properties.ListCategories)

	mux.HandleFunc("GET /api/v1/analytics/pricing-trends", analytics.GetPricingTrends)

	mux.HandleFunc("GET /api/v1/compare", search.GetComparison)
	mux.HandleFunc("POST /api/v1/compare", search.AddToComparison)
	mux.HandleFunc("DELETE /api/v1/compare/{propertyId}", search.RemoveFromComparison)
	mux.HandleFunc("GET /api/v1/compare/results", search.GetComparisonResults)

	mux.HandleFunc("POST /api/v1/payments/webhooks/mtn", payments.MtnWebhook)
	mux.HandleFunc("POST /api/v1/payments/webhooks/airtel", payments.AirtelWebhook)
	mux.HandleFunc("POST /api/v1/payments/webhooks/zamtel", payments.ZamtelWebhook)

	mux.HandleFunc("GET /api/v1/subscriptions/plans", payments.GetPlans)

	// ==========================================
	// 2. AUTHENTICATED GENERAL ROUTES
	// ==========================================
	authHandler := func(pattern string, handlerFunc http.HandlerFunc) {
		mux.Handle(pattern, chain(handlerFunc, authMW))
	}

	roleHandler := func(pattern string, handlerFunc http.HandlerFunc, roleGuard Middleware) {
		mux.Handle(pattern, chain(handlerFunc, authMW, roleGuard))
	}

	authHandler("POST /api/v1/auth/logout", auth.Logout)
	authHandler("POST /api/v1/auth/mfa/enable", auth.EnableMFA)
	authHandler("POST /api/v1/auth/mfa/disable", auth.DisableMFA)
	authHandler("GET /api/v1/auth/sessions", auth.ListSessions)
	authHandler("DELETE /api/v1/auth/sessions/{id}", auth.RevokeSession)

	authHandler("GET /api/v1/users/me", users.GetMe)
	authHandler("PUT /api/v1/users/me", users.UpdateMe)
	authHandler("DELETE /api/v1/users/me", users.DeactivateMe)
	authHandler("GET /api/v1/users/{id}", users.GetUserProfile)
	authHandler("PUT /api/v1/users/{id}/avatar", users.UpdateAvatar)
	authHandler("GET /api/v1/users/tenants/{id}", users.GetTenantProfile)
	authHandler("PUT /api/v1/users/tenants/{id}", users.UpdateTenantProfile)
	authHandler("GET /api/v1/users/landlords/{id}", users.GetLandlordProfile)
	authHandler("PUT /api/v1/users/landlords/{id}", users.UpdateLandlordProfile)
	authHandler("GET /api/v1/users/agents/{id}", users.GetAgentProfile)
	authHandler("PUT /api/v1/users/agents/{id}", users.UpdateAgentProfile)
	authHandler("GET /api/v1/users/{id}/preferences", users.GetPreferences)
	authHandler("PUT /api/v1/users/{id}/preferences", users.UpdatePreferences)

	authHandler("POST /api/v1/properties/{id}/reports", properties.ReportListing)
	authHandler("GET /api/v1/properties/{id}/history", properties.GetHistory)

	roleHandler("GET /api/v1/favorites", favorites.ListFavorites, roleTenant)
	roleHandler("POST /api/v1/favorites", favorites.AddFavorite, roleTenant)
	roleHandler("DELETE /api/v1/favorites/{propertyId}", favorites.RemoveFavorite, roleTenant)
	roleHandler("GET /api/v1/saved-searches", favorites.ListSavedSearches, roleTenant)
	roleHandler("POST /api/v1/saved-searches", favorites.SaveSearch, roleTenant)
	roleHandler("DELETE /api/v1/saved-searches/{id}", favorites.DeleteSavedSearch, roleTenant)

	authHandler("GET /api/v1/inquiries", inquiries.ListInquiries)
	authHandler("POST /api/v1/inquiries", inquiries.CreateInquiry)
	authHandler("GET /api/v1/inquiries/{id}", inquiries.GetInquiry)
	authHandler("PATCH /api/v1/inquiries/{id}/close", inquiries.CloseInquiry)
	authHandler("POST /api/v1/inquiries/{id}/replies", inquiries.ReplyInquiry)

	authHandler("GET /api/v1/viewings", inquiries.ListViewings)
	authHandler("POST /api/v1/viewings", inquiries.CreateViewing)
	authHandler("GET /api/v1/viewings/{id}", inquiries.GetViewing)
	authHandler("PATCH /api/v1/viewings/{id}/reschedule", inquiries.RescheduleViewing)
	authHandler("POST /api/v1/viewings/{id}/feedback", inquiries.SubmitViewingFeedback)

	authHandler("GET /api/v1/bookings", inquiries.ListBookings)
	authHandler("POST /api/v1/bookings", inquiries.CreateBooking)
	authHandler("GET /api/v1/bookings/{id}", inquiries.GetBooking)
	authHandler("PATCH /api/v1/bookings/{id}/cancel", inquiries.CancelBooking)

	authHandler("GET /api/v1/leases", inquiries.ListLeases)
	authHandler("GET /api/v1/leases/{id}", inquiries.GetLease)
	authHandler("PATCH /api/v1/leases/{id}/sign", inquiries.SignLease)
	authHandler("POST /api/v1/leases/{id}/terminate", inquiries.TerminateLease)

	authHandler("GET /api/v1/screening", inquiries.ListScreenings)
	authHandler("GET /api/v1/screening/{id}", inquiries.GetScreening)

	authHandler("POST /api/v1/payments/mobile-money", payments.InitiateMobileMoney)
	authHandler("GET /api/v1/payments/transactions", payments.ListTransactions)
	authHandler("GET /api/v1/payments/transactions/{id}", payments.GetTransaction)
	authHandler("GET /api/v1/payments/refunds", payments.ListRefunds)
	authHandler("POST /api/v1/payments/refunds", payments.RequestRefund)
	authHandler("GET /api/v1/payments/refunds/{id}", payments.GetRefund)

	authHandler("GET /api/v1/subscriptions/my", payments.GetMySubscription)
	authHandler("POST /api/v1/subscriptions/subscribe", payments.Subscribe)
	authHandler("POST /api/v1/subscriptions/cancel", payments.CancelSubscription)

	authHandler("GET /api/v1/ws/connect", ws.Connect)
	authHandler("GET /api/v1/conversations", ws.ListConversations)
	authHandler("POST /api/v1/conversations", ws.CreateConversation)
	authHandler("GET /api/v1/conversations/{id}", ws.GetConversation)
	authHandler("GET /api/v1/conversations/{id}/messages", ws.ListMessages)
	authHandler("POST /api/v1/conversations/{id}/messages/http", ws.SendMessageHTTP)
	authHandler("DELETE /api/v1/conversations/{id}/messages/{msgId}", ws.DeleteMessage)
	authHandler("GET /api/v1/notifications", ws.ListNotifications)
	authHandler("PATCH /api/v1/notifications/{id}/read", ws.MarkNotificationRead)
	authHandler("PATCH /api/v1/notifications/read-all", ws.MarkAllNotificationsRead)
	authHandler("DELETE /api/v1/notifications/{id}", ws.DeleteNotification)

	authHandler("POST /api/v1/uploads/presigned-url", uploads.GeneratePresignedURL)
	authHandler("POST /api/v1/uploads/file", uploads.DirectUpload)
	authHandler("GET /api/v1/uploads/my-files", uploads.ListMyFiles)
	authHandler("DELETE /api/v1/uploads/{fileId}", uploads.DeleteFile)

	// ==========================================
	// 3. LANDLORD / AGENT PROTECTED ROUTES
	// ==========================================


	roleHandler("POST /api/v1/properties", properties.Create, roleAgentLandlord)
	roleHandler("PUT /api/v1/properties/{id}", properties.Update, roleAgentLandlord)
	roleHandler("DELETE /api/v1/properties/{id}", properties.Delete, roleAgentLandlord)
	roleHandler("PATCH /api/v1/properties/{id}/status", properties.UpdateStatus, roleAgentLandlord)
	roleHandler("POST /api/v1/properties/{id}/images", properties.UploadImages, roleAgentLandlord)
	roleHandler("DELETE /api/v1/properties/{id}/images/{imageId}", properties.DeleteImage, roleAgentLandlord)
	roleHandler("POST /api/v1/properties/{id}/availability", properties.AddAvailabilityBlock, roleAgentLandlord)
	roleHandler("POST /api/v1/properties/{id}/amenities", properties.UpdateAmenities, roleAgentLandlord)

	roleHandler("PATCH /api/v1/viewings/{id}/approve", inquiries.ApproveViewing, roleAgentLandlord)
	roleHandler("PATCH /api/v1/viewings/{id}/reject", inquiries.RejectViewing, roleAgentLandlord)

	roleHandler("PATCH /api/v1/bookings/{id}/approve", inquiries.ApproveBooking, roleAgentLandlord)
	roleHandler("POST /api/v1/leases", inquiries.CreateLease, roleAgentLandlord)
	roleHandler("POST /api/v1/screening", inquiries.CreateScreening, roleAgentLandlord)

	roleHandler("GET /api/v1/analytics/properties/{id}", analytics.GetPropertyAnalytics, roleAgentLandlord)
	roleHandler("GET /api/v1/analytics/landlord/dashboard", analytics.GetLandlordDashboard, roleLandlord)

	// ==========================================
	// 4. AGENT ONLY ROUTES
	// ==========================================
	roleHandler("GET /api/v1/crm/leads", crm.ListLeads, roleAgent)
	roleHandler("POST /api/v1/crm/leads", crm.CreateLead, roleAgent)
	roleHandler("GET /api/v1/crm/leads/{id}", crm.GetLead, roleAgent)
	roleHandler("PUT /api/v1/crm/leads/{id}", crm.UpdateLead, roleAgent)
	roleHandler("POST /api/v1/crm/leads/{id}/notes", crm.AddLeadNote, roleAgent)

	roleHandler("GET /api/v1/crm/tasks", crm.ListTasks, roleAgent)
	roleHandler("POST /api/v1/crm/tasks", crm.CreateTask, roleAgent)
	roleHandler("PUT /api/v1/crm/tasks/{id}", crm.UpdateTask, roleAgent)
	roleHandler("DELETE /api/v1/crm/tasks/{id}", crm.DeleteTask, roleAgent)

	roleHandler("GET /api/v1/crm/deals", crm.GetDeals, roleAgent)
	roleHandler("GET /api/v1/crm/meetings", crm.ListMeetings, roleAgent)
	roleHandler("POST /api/v1/crm/meetings", crm.CreateMeeting, roleAgent)
	roleHandler("GET /api/v1/crm/performance", crm.GetPerformance, roleAgent)

	// ==========================================
	// 5. ADMIN / MODERATION ROUTES
	// ==========================================
	roleHandler("GET /api/v1/admin/users", admin.ListUsers, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/users/{id}/role", admin.UpdateUserRole, roleAdminOnly)
	roleHandler("PATCH /api/v1/admin/users/{id}/ban", admin.UpdateUserBan, roleAdminMod)

	roleHandler("GET /api/v1/admin/moderation/properties", admin.ListModerationProperties, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/moderation/properties/{id}/approve", admin.ApproveProperty, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/moderation/properties/{id}/reject", admin.RejectProperty, roleAdminMod)

	roleHandler("GET /api/v1/admin/moderation/reports", admin.ListReports, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/moderation/reports/{id}/resolve", admin.ResolveReport, roleAdminMod)

	roleHandler("GET /api/v1/admin/verifications", admin.ListVerifications, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/verifications/{id}/approve", admin.ApproveVerification, roleAdminMod)
	roleHandler("PATCH /api/v1/admin/verifications/{id}/reject", admin.RejectVerification, roleAdminMod)

	roleHandler("GET /api/v1/admin/audit-logs", admin.ListAuditLogs, roleAdminOnly)
	roleHandler("GET /api/v1/admin/settings", admin.GetSystemSettings, roleAdminOnly)
	roleHandler("PUT /api/v1/admin/settings", admin.UpdateSystemSettings, roleAdminOnly)

	roleHandler("GET /api/v1/analytics/platform/overview", analytics.GetPlatformOverview, roleAdminOnly)
	roleHandler("GET /api/v1/analytics/platform/revenue", analytics.GetPlatformRevenue, roleAdminOnly)

	roleHandler("POST /api/v1/properties/categories", properties.CreateCategory, roleAdminOnly)
	roleHandler("PUT /api/v1/properties/categories/{id}", properties.UpdateCategory, roleAdminOnly)
	roleHandler("DELETE /api/v1/properties/categories/{id}", properties.DeleteCategory, roleAdminOnly)

	roleHandler("POST /api/v1/search-index/reindex", search.Reindex, roleAdminOnly)
	roleHandler("GET /api/v1/search-index/status", search.GetIndexStatus, roleAdminOnly)
	roleHandler("GET /api/v1/search-index/synonyms", search.ListSynonyms, roleAdminOnly)
	roleHandler("POST /api/v1/search-index/synonyms", search.CreateSynonym, roleAdminOnly)
	roleHandler("DELETE /api/v1/search-index/synonyms/{id}", search.DeleteSynonym, roleAdminOnly)

	// Wrap mux with globally standard recovery and logging
	return chain(mux, middleware.Recover, middleware.Logger)
}
