package handlers

import (
	"encoding/json"
	"net/http"
)

type UploadsHandler struct{}

func NewUploadsHandler() *UploadsHandler {
	return &UploadsHandler{}
}

func (h *UploadsHandler) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"upload_url": "https://home-rent-storage.s3.amazonaws.com/uploads/" + body.Filename + "?AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Signature=vjbyPxybdZaNmGa%2ByT272YEAiv4%3D&Expires=1775700000",
		"file_url":   "https://cdn.site.com/uploads/" + body.Filename,
	})
}

func (h *UploadsHandler) DirectUpload(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         "img_uploaded_8829",
		"url":        "https://cdn.site.com/uploads/direct_file.jpg",
		"created_at": "2026-06-08T17:55:00Z",
	})
}

func (h *UploadsHandler) ListMyFiles(w http.ResponseWriter, r *http.Request) {
	files := []map[string]interface{}{
		{
			"id":         "img_uploaded_8829",
			"url":        "https://cdn.site.com/uploads/direct_file.jpg",
			"created_at": "2026-06-08T17:55:00Z",
		},
	}
	json.NewEncoder(w).Encode(files)
}

func (h *UploadsHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
