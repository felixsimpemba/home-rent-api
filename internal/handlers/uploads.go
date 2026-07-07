package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierrors "github.com/felixsimpemba/home-rent-api/internal/errors"
	"github.com/felixsimpemba/home-rent-api/internal/middleware"
	"github.com/felixsimpemba/home-rent-api/internal/models"
	"github.com/felixsimpemba/home-rent-api/internal/respond"
)

// UploadsHandler handles file upload endpoints
type UploadsHandler struct {
	db *gorm.DB
}

// NewUploadsHandler creates an UploadsHandler with DB
func NewUploadsHandler(db *gorm.DB) *UploadsHandler {
	return &UploadsHandler{db: db}
}

// GeneratePresignedURL simulates presigned URL generation (TODO: integrate S3/MinIO)
func (h *UploadsHandler) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filename string `json:"filename"`
		MimeType string `json:"mime_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Filename == "" {
		apierrors.BadRequest(w, r, "filename is required.", nil)
		return
	}

	fileID := "file_" + uuid.NewString()
	// TODO: Generate actual presigned URL from S3/MinIO/Cloudflare R2
	presignedURL := "https://storage.homerent.zm/uploads/" + fileID + "/" + body.Filename

	respond.OK(w, map[string]interface{}{
		"file_id":      fileID,
		"upload_url":   presignedURL,
		"expires_in":   900, // 15 minutes
		"method":       "PUT",
	})
}

// DirectUpload handles multipart file upload directly to the server
func (h *UploadsHandler) DirectUpload(w http.ResponseWriter, r *http.Request) {
	uploaderID := middleware.GetUserID(r)

	// 10MB max size
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		apierrors.BadRequest(w, r, "Failed to parse multipart form. Max size is 10MB.", nil)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		apierrors.BadRequest(w, r, "No file found in request. Use 'file' as the form field name.", nil)
		return
	}
	defer file.Close()

	fileID := "file_" + uuid.NewString()
	// TODO: Actually store the file to disk/cloud storage
	// For now record metadata in DB
	uploadedFile := models.UploadedFile{
		ID:         fileID,
		UploaderID: uploaderID,
		Filename:   handler.Filename,
		URL:        "https://storage.homerent.zm/uploads/" + fileID + "/" + handler.Filename,
		MimeType:   handler.Header.Get("Content-Type"),
		SizeBytes:  handler.Size,
	}
	h.db.Create(&uploadedFile)

	respond.Created201(w, map[string]interface{}{
		"file_id":  uploadedFile.ID,
		"filename": uploadedFile.Filename,
		"url":      uploadedFile.URL,
		"size":     uploadedFile.SizeBytes,
	})
}

// ListMyFiles returns all files uploaded by the authenticated user
func (h *UploadsHandler) ListMyFiles(w http.ResponseWriter, r *http.Request) {
	uploaderID := middleware.GetUserID(r)
	page, limit := pageLimit(r)

	var total int64
	h.db.Model(&models.UploadedFile{}).Where("uploader_id = ?", uploaderID).Count(&total)

	var files []models.UploadedFile
	h.db.Where("uploader_id = ?", uploaderID).
		Offset((page - 1) * limit).Limit(limit).
		Order("created_at DESC").Find(&files)

	respond.Paginated(w, files, page, limit, total)
}

// DeleteFile removes a file record (and should delete from storage in production)
func (h *UploadsHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	uploaderID := middleware.GetUserID(r)
	fileID := r.PathValue("fileId")

	var file models.UploadedFile
	if err := h.db.First(&file, "id = ? AND uploader_id = ?", fileID, uploaderID).Error; err != nil {
		apierrors.NotFound(w, r, "File not found or access denied.")
		return
	}

	// TODO: Delete from actual cloud storage
	h.db.Delete(&file)
	respond.NoContent(w)
}
