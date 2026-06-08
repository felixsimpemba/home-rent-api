package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/felixsimpemba/home-rent-api/internal/errors"
)

type contextKey string

const (
	UserIDKey contextKey = "userId"
	UserRoleKey contextKey = "userRole"
)

// Logger logs incoming requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// Simple response wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Printf("Completed %s %s with %d %s in %v",
			r.Method, r.URL.Path, wrapped.status, http.StatusText(wrapped.status), time.Since(start))
	})
}

// Recover handles panics gracefully
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

// Authenticate extracts claims from JWT Bearer token and populates context
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Allow request to proceed unauthenticated; handlers will enforce auth if needed
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errors.Unauthorized(w, r, "Invalid authorization format. Must be 'Bearer <token>'.")
				return
			}

			token := parts[1]
			// Mock verification: In a real system, verify token using jwt-go and secret
			// For this boilerplate, we accept 'tenant_token', 'landlord_token', 'agent_token', 'admin_token'
			var userID, role string
			switch token {
			case "tenant_token":
				userID = "usr_tenant_123"
				role = "tenant"
			case "landlord_token":
				userID = "usr_landlord_456"
				role = "landlord"
			case "agent_token":
				userID = "usr_agent_789"
				role = "agent"
			case "admin_token":
				userID = "usr_admin_000"
				role = "admin"
			default:
				// Fallback to checking if token matches direct roles for ease of mocking/testing
				if strings.HasSuffix(token, "_token") {
					role = strings.TrimSuffix(token, "_token")
					userID = "usr_mock_" + role
				} else {
					errors.Unauthorized(w, r, "Invalid or expired token.")
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles restricts path access to specified roles
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || role == "" {
				errors.Unauthorized(w, r, "Authentication required to access this resource.")
				return
			}

			allowed := false
			for _, r := range allowedRoles {
				if r == role {
					allowed = true
					break
				}
			}

			if !allowed {
				errors.Forbidden(w, r, "You do not have the required permissions to access this resource.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
