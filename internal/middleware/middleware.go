package middleware

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/felixsimpemba/home-rent-api/internal/errors"
)

type contextKey string

const (
	UserIDKey   contextKey = "userId"
	UserRoleKey contextKey = "userRole"
)

// ─── Logger ───────────────────────────────────────────────────────────────────

// Logger logs incoming requests and completion status
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and log request body for non-GET/non-DELETE/non-OPTIONS requests
		var bodyStr string
		if r.Body != nil && r.Method != http.MethodGet && r.Method != http.MethodDelete && r.Method != http.MethodOptions {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				// Restore body
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(bodyBytes) > 0 {
					bodyStr = string(bodyBytes)
					if len(bodyStr) > 1000 {
						bodyStr = bodyStr[:1000] + "... (truncated)"
					}
				}
			}
		}

		if bodyStr != "" {
			log.Printf("→ %s %s | Data: %s", r.Method, r.URL.RequestURI(), bodyStr)
		} else {
			log.Printf("→ %s %s", r.Method, r.URL.RequestURI())
		}

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Printf("← %s %s %d %s (%v)",
			r.Method, r.URL.RequestURI(), wrapped.status, http.StatusText(wrapped.status), time.Since(start))
	})
}

// ─── Recover ──────────────────────────────────────────────────────────────────

// Recover handles panics gracefully and returns 500
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				errors.InternalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ─── Authenticate ─────────────────────────────────────────────────────────────

// Claims represents JWT payload
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Authenticate validates a JWT Bearer token and injects user context
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.Unauthorized(w, r, "Authorization header is required.")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				errors.Unauthorized(w, r, "Invalid authorization format. Expected: Bearer <token>")
				return
			}

			tokenStr := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				errors.Unauthorized(w, r, "Invalid or expired access token.")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ─── RequireRoles ─────────────────────────────────────────────────────────────

// RequireRoles restricts access to the specified roles
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || role == "" {
				errors.Unauthorized(w, r, "Authentication required to access this resource.")
				return
			}

			for _, allowed := range allowedRoles {
				if allowed == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			errors.Forbidden(w, r, "You do not have the required permissions.")
		})
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// GetUserID extracts the authenticated user ID from request context
func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(UserIDKey).(string)
	return id
}

// GetUserRole extracts the authenticated user role from request context
func GetUserRole(r *http.Request) string {
	role, _ := r.Context().Value(UserRoleKey).(string)
	return role
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}
