package respond

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// OK sends a 200 JSON response
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created sends a 201 JSON response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 response with no body
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Success wraps data in a standard success envelope
func Success(w http.ResponseWriter, data interface{}) {
	OK(w, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// Created201 wraps data in a standard success envelope with 201 status
func Created201(w http.ResponseWriter, data interface{}) {
	Created(w, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// Message sends a simple success message response
func Message(w http.ResponseWriter, msg string) {
	OK(w, map[string]interface{}{
		"success": true,
		"message": msg,
	})
}

// Paginated sends a paginated list response
func Paginated(w http.ResponseWriter, data interface{}, page, limit int, total int64) {
	OK(w, map[string]interface{}{
		"data": data,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
