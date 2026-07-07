package errors

import (
	"encoding/json"
	"net/http"
)

// InvalidParam represents a single field validation error
type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ErrorResponse represents an RFC 7807 Problem Details response
type ErrorResponse struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}

// WriteJSONError sends a formatted problem details response to the client
func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, title, detail string, invalidParams []InvalidParam) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	errResponse := ErrorResponse{
		Type:          "https://api.homerent.zm/errors/type-uri",
		Title:         title,
		Status:        status,
		Detail:        detail,
		Instance:      r.URL.Path,
		InvalidParams: invalidParams,
	}

	_ = json.NewEncoder(w).Encode(errResponse)
}

// Helper methods for common errors
func BadRequest(w http.ResponseWriter, r *http.Request, detail string, invalidParams []InvalidParam) {
	WriteJSONError(w, r, http.StatusBadRequest, "Bad Request", detail, invalidParams)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	WriteJSONError(w, r, http.StatusUnauthorized, "Unauthorized", detail, nil)
}

func Forbidden(w http.ResponseWriter, r *http.Request, detail string) {
	WriteJSONError(w, r, http.StatusForbidden, "Forbidden", detail, nil)
}

func NotFound(w http.ResponseWriter, r *http.Request, detail string) {
	WriteJSONError(w, r, http.StatusNotFound, "Not Found", detail, nil)
}

func Conflict(w http.ResponseWriter, r *http.Request, detail string) {
	WriteJSONError(w, r, http.StatusConflict, "Conflict", detail, nil)
}

func UnprocessableEntity(w http.ResponseWriter, r *http.Request, detail string, params []InvalidParam) {
	WriteJSONError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", detail, params)
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, r, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred on the server.", nil)
}
