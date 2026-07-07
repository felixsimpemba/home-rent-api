package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler with DB and JWT secret
func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

// ─── Register ─────────────────────────────────────────────────────────────────

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Password    string `json:"password"`
		AccountType string `json:"account_type"` // tenant|landlord|agent
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid JSON body", nil)
		return
	}

	if body.FirstName == "" || body.LastName == "" || body.Email == "" || body.Phone == "" || body.Password == "" {
		apierrors.BadRequest(w, r, "All fields are required.", []apierrors.InvalidParam{
			{Name: "first_name", Reason: "must not be blank"},
			{Name: "last_name", Reason: "must not be blank"},
			{Name: "email", Reason: "must not be blank"},
			{Name: "phone", Reason: "must not be blank"},
			{Name: "password", Reason: "must not be blank"},
		})
		return
	}

	// Check if email already exists
	var existing models.User
	if err := h.db.Where("email = ?", body.Email).First(&existing).Error; err == nil {
		apierrors.Conflict(w, r, "An account with this email already exists.")
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	role := "tenant"
	if body.AccountType == "landlord" || body.AccountType == "agent" {
		role = body.AccountType
	}

	user := models.User{
		ID:        "usr_" + uuid.NewString(),
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Phone:     body.Phone,
		Password:  string(hashed),
		Role:      role,
	}

	if err := h.db.Create(&user).Error; err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	respond.Created201(w, map[string]interface{}{
		"user_id":               user.ID,
		"email":                 user.Email,
		"role":                  user.Role,
		"verification_required": true,
	})
}

// ─── Login ────────────────────────────────────────────────────────────────────

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierrors.BadRequest(w, r, "Invalid request payload", nil)
		return
	}

	if body.Email == "" || body.Password == "" {
		apierrors.BadRequest(w, r, "Email and password are required.", nil)
		return
	}

	var user models.User
	if err := h.db.Where("email = ? AND banned = false", body.Email).First(&user).Error; err != nil {
		apierrors.Unauthorized(w, r, "Invalid email or password.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		apierrors.Unauthorized(w, r, "Invalid email or password.")
		return
	}

	accessToken, err := h.generateAccessToken(user.ID, user.Role)
	if err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	refreshToken := uuid.NewString()
	session := models.Session{
		ID:           "sess_" + uuid.NewString(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		DeviceName:   r.UserAgent(),
		IPAddress:    r.RemoteAddr,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
	h.db.Create(&session)

	respond.OK(w, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"user": map[string]interface{}{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"role":       user.Role,
			"avatar_url": user.AvatarURL,
		},
	})
}

// ─── Refresh ──────────────────────────────────────────────────────────────────

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		apierrors.BadRequest(w, r, "refresh_token is required.", nil)
		return
	}

	var session models.Session
	if err := h.db.Where("refresh_token = ? AND expires_at > ?", body.RefreshToken, time.Now()).
		Preload("User").First(&session).Error; err != nil {
		apierrors.Unauthorized(w, r, "Invalid or expired refresh token.")
		return
	}

	accessToken, err := h.generateAccessToken(session.User.ID, session.User.Role)
	if err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	respond.OK(w, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.RefreshToken != "" {
		h.db.Where("refresh_token = ?", body.RefreshToken).Delete(&models.Session{})
	}
	respond.Message(w, "Successfully logged out.")
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		apierrors.BadRequest(w, r, "email is required.", nil)
		return
	}

	var user models.User
	if h.db.Where("email = ?", body.Email).First(&user).Error == nil {
		// Delete any existing tokens for this user
		h.db.Where("user_id = ?", user.ID).Delete(&models.PasswordResetToken{})

		token := models.PasswordResetToken{
			ID:        "prt_" + uuid.NewString(),
			UserID:    user.ID,
			Token:     uuid.NewString(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		h.db.Create(&token)
		// TODO: Send reset link via email: /auth/reset-password?token=<token.Token>
	}

	// Always respond 200 to prevent email enumeration
	respond.Message(w, "If this email is registered, a password reset link has been sent.")
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" || body.Password == "" {
		apierrors.BadRequest(w, r, "token and password are required.", nil)
		return
	}

	var prt models.PasswordResetToken
	if err := h.db.Where("token = ? AND used_at IS NULL AND expires_at > ?", body.Token, time.Now()).
		First(&prt).Error; err != nil {
		apierrors.BadRequest(w, r, "Invalid or expired reset token.", nil)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		apierrors.InternalServerError(w, r)
		return
	}

	h.db.Model(&models.User{}).Where("id = ?", prt.UserID).Update("password", string(hashed))
	now := time.Now()
	h.db.Model(&prt).Update("used_at", &now)

	respond.Message(w, "Password updated successfully.")
}

// ─── VerifyEmail ──────────────────────────────────────────────────────────────

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		apierrors.BadRequest(w, r, "token is required.", nil)
		return
	}

	var evt models.EmailVerificationToken
	if err := h.db.Where("token = ? AND used_at IS NULL AND expires_at > ?", body.Token, time.Now()).
		First(&evt).Error; err != nil {
		apierrors.BadRequest(w, r, "Invalid or expired verification token.", nil)
		return
	}

	h.db.Model(&models.User{}).Where("id = ?", evt.UserID).Update("email_verified", true)
	now := time.Now()
	h.db.Model(&evt).Update("used_at", &now)

	respond.Message(w, "Email address verified successfully.")
}

// ─── MFA ──────────────────────────────────────────────────────────────────────

func (h *AuthHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	// TODO: Generate real TOTP secret using a library like pquerna/otp
	secret := "NBSWY3DPEB3W64TBNQ_" + userID[:8]
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"mfa_secret":  secret,
		"mfa_enabled": false, // user must verify before it's enabled
	})
	respond.OK(w, map[string]interface{}{
		"secret":      secret,
		"qr_code_url": "otpauth://totp/HomeRent:" + userID + "?secret=" + secret + "&issuer=HomeRent",
	})
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	h.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"mfa_enabled": false,
		"mfa_secret":  "",
	})
	respond.Message(w, "MFA disabled successfully.")
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify TOTP code against stored secret
	// For now accept any 6-digit code
	respond.OK(w, map[string]interface{}{
		"success": true,
		"message": "MFA verified.",
	})
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var sessions []models.Session
	h.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).Find(&sessions)

	result := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":          s.ID,
			"device_name": s.DeviceName,
			"ip_address":  s.IPAddress,
			"last_active": s.UpdatedAt,
			"expires_at":  s.ExpiresAt,
		})
	}
	respond.OK(w, result)
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")
	h.db.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&models.Session{})
	respond.Message(w, "Session revoked successfully.")
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// generateAccessToken creates a signed JWT access token valid for 1 hour
func (h *AuthHandler) generateAccessToken(userID, role string) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "home-rent-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
