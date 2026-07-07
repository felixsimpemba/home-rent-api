package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// PaymentsHandler handles payment and subscription endpoints
type PaymentsHandler struct {
	db *gorm.DB
}

// NewPaymentsHandler creates a PaymentsHandler with DB
func NewPaymentsHandler(db *gorm.DB) *PaymentsHandler {
	return &PaymentsHandler{db: db}
}

// ─── Mobile Money ─────────────────────────────────────────────────────────────

func (h *PaymentsHandler) InitiateMobileMoney(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		Provider    string  `json:"provider"` // mtn|airtel|zamtel
		PhoneNumber string  `json:"phone_number"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Purpose     string  `json:"purpose"` // rent|deposit|subscription
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}
	if body.Provider == "" || body.PhoneNumber == "" || body.Amount <= 0 {
		apierrors.BadRequest(w, r, "provider, phone_number, and amount are required.", nil)
		return
	}
	if body.Currency == "" {
		body.Currency = "ZMW"
	}

	ref := "TXN-" + uuid.NewString()[:8]
	tx := models.Transaction{
		ID:          "txn_" + uuid.NewString(),
		Reference:   ref,
		UserID:      userID,
		Provider:    body.Provider,
		PhoneNumber: body.PhoneNumber,
		Amount:      body.Amount,
		Currency:    body.Currency,
		Purpose:     body.Purpose,
		Status:      "pending",
	}
	h.db.Create(&tx)

	// TODO: Call actual MTN/Airtel/Zamtel MoMo API here and update ProviderRef
	// For now the transaction is "pending" — webhook will flip it to "successful"

	respond.Created201(w, map[string]interface{}{
		"transaction_id": tx.ID,
		"reference":      tx.Reference,
		"status":         tx.Status,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"provider":       tx.Provider,
		"message":        "Payment initiated. Waiting for confirmation.",
	})
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

func (h *PaymentsHandler) processWebhook(w http.ResponseWriter, r *http.Request, provider string) {
	var body struct {
		Reference string `json:"reference"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid webhook payload", nil)
		return
	}

	result := h.db.Model(&models.Transaction{}).
		Where("reference = ? AND provider = ?", body.Reference, provider).
		Update("status", body.Status)

	if result.RowsAffected == 0 {
		apierrors.NotFound(w, r, "Transaction not found for this reference.")
		return
	}

	respond.Message(w, "Webhook processed successfully.")
}

func (h *PaymentsHandler) MtnWebhook(w http.ResponseWriter, r *http.Request) {
	h.processWebhook(w, r, "mtn")
}

func (h *PaymentsHandler) AirtelWebhook(w http.ResponseWriter, r *http.Request) {
	h.processWebhook(w, r, "airtel")
}

func (h *PaymentsHandler) ZamtelWebhook(w http.ResponseWriter, r *http.Request) {
	h.processWebhook(w, r, "zamtel")
}

// ─── Transactions ─────────────────────────────────────────────────────────────

func (h *PaymentsHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Transaction{})
	if role != "admin" {
		tx = tx.Where("user_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var transactions []models.Transaction
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&transactions)
	respond.Paginated(w, transactions, page, limit, total)
}

func (h *PaymentsHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	id := r.PathValue("id")

	var txn models.Transaction
	if err := h.db.First(&txn, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Transaction not found.")
		return
	}
	if role != "admin" && txn.UserID != userID {
		apierrors.Forbidden(w, r, "Access denied.")
		return
	}
	respond.OK(w, txn)
}

// ─── Refunds ──────────────────────────────────────────────────────────────────

func (h *PaymentsHandler) ListRefunds(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	page, limit := pageLimit(r)

	tx := h.db.Model(&models.Refund{}).Preload("Transaction")
	if role != "admin" {
		tx = tx.Where("user_id = ?", userID)
	}

	var total int64
	tx.Count(&total)
	var refunds []models.Refund
	tx.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&refunds)
	respond.Paginated(w, refunds, page, limit, total)
}

func (h *PaymentsHandler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		TransactionID string  `json:"transaction_id"`
		Reason        string  `json:"reason"`
		Amount        float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TransactionID == "" || body.Reason == "" {
		apierrors.BadRequest(w, r, "transaction_id and reason are required.", nil)
		return
	}

	var txn models.Transaction
	if err := h.db.First(&txn, "id = ? AND user_id = ?", body.TransactionID, userID).Error; err != nil {
		apierrors.NotFound(w, r, "Transaction not found.")
		return
	}

	refundAmount := body.Amount
	if refundAmount <= 0 {
		refundAmount = txn.Amount
	}

	refund := models.Refund{
		ID:            "ref_" + uuid.NewString(),
		TransactionID: body.TransactionID,
		UserID:        userID,
		Reason:        body.Reason,
		Amount:        refundAmount,
		Status:        "pending",
	}
	h.db.Create(&refund)
	respond.Created201(w, refund)
}

func (h *PaymentsHandler) GetRefund(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var refund models.Refund
	if err := h.db.Preload("Transaction").First(&refund, "id = ?", id).Error; err != nil {
		apierrors.NotFound(w, r, "Refund not found.")
		return
	}
	respond.OK(w, refund)
}

// ─── Subscription Plans ───────────────────────────────────────────────────────

func (h *PaymentsHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	var plans []models.SubscriptionPlan
	h.db.Where("active = true").Find(&plans)

	// Seed default plans if empty
	if len(plans) == 0 {
		defaultPlans := []models.SubscriptionPlan{
			{
				ID:           "plan_free",
				Name:         "Free",
				Description:  "Basic listing access",
				PriceMonthly: 0,
				Features:     `["5 listings","Basic search","Email support"]`,
				Active:       true,
			},
			{
				ID:           "plan_pro",
				Name:         "Pro Landlord",
				Description:  "For active landlords",
				PriceMonthly: 299,
				Features:     `["Unlimited listings","Featured listings","Analytics","Priority support"]`,
				Active:       true,
			},
			{
				ID:           "plan_agent",
				Name:         "Agent Plus",
				Description:  "For professional agents",
				PriceMonthly: 599,
				Features:     `["Unlimited listings","CRM tools","Lead management","Advanced analytics","Dedicated support"]`,
				Active:       true,
			},
		}
		h.db.Create(&defaultPlans)
		plans = defaultPlans
	}

	respond.OK(w, plans)
}

func (h *PaymentsHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var sub models.Subscription
	if err := h.db.Preload("Plan").Where("user_id = ? AND status = ?", userID, "active").
		First(&sub).Error; err != nil {
		respond.OK(w, map[string]interface{}{
			"has_subscription": false,
			"plan":             nil,
		})
		return
	}
	respond.OK(w, map[string]interface{}{
		"has_subscription": true,
		"subscription":     sub,
	})
}

func (h *PaymentsHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		PlanID string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PlanID == "" {
		apierrors.BadRequest(w, r, "plan_id is required.", nil)
		return
	}

	var plan models.SubscriptionPlan
	if err := h.db.First(&plan, "id = ?", body.PlanID).Error; err != nil {
		apierrors.NotFound(w, r, "Subscription plan not found.")
		return
	}

	// Cancel existing subscription
	h.db.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").
		Update("status", "cancelled")

	sub := models.Subscription{
		ID:        "sub_" + uuid.NewString(),
		UserID:    userID,
		PlanID:    body.PlanID,
		Status:    "active",
		StartedAt: time.Now(),
	}
	h.db.Create(&sub)
	h.db.Preload("Plan").First(&sub, "id = ?", sub.ID)
	respond.Created201(w, sub)
}

func (h *PaymentsHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	h.db.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").
		Update("status", "cancelled")
	respond.Message(w, "Subscription cancelled successfully.")
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func nowTime() time.Time {
	return time.Now()
}
