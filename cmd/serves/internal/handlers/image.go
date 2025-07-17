package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/imageutil"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/store"
)

// ImageUploadResponse represents the response for image upload operations
type ImageUploadResponse struct {
	ImageURL string `json:"image_url"`
	Message  string `json:"message"`
}

// PresignedUploadResponse represents the response for presigned upload requests
type PresignedUploadResponse struct {
	UploadURL string `json:"upload_url"`
	ImageKey  string `json:"image_key"`
	Message   string `json:"message"`
}

// UploadHomeImage handles direct image upload for homes
func UploadHomeImage(s store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("home-id is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			//rest.ErrUnauthorized(w, errors.New("user ID is required"))
			userID = "test-user" // For testing purposes, remove in production
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			rest.ErrBadRequest(w, fmt.Errorf("failed to parse multipart form: %w", err))
			return
		}

		// Get the file from form
		file, header, err := r.FormFile("image")
		if err != nil {
			rest.ErrBadRequest(w, fmt.Errorf("failed to get image file: %w", err))
			return
		}
		defer file.Close()

		// Validate content type
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = imageutil.ExtractContentTypeFromFilename(header.Filename)
		}

		if err := imageutil.ValidateImageType(contentType); err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Generate unique key for the image
		imageKey := imageutil.GenerateImageKey(userID, homeID, contentType)

		// Upload to storage
		imageURL, err := s.Upload(r.Context(), imageKey, file, contentType)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to upload image: %w", err))
			return
		}

		rest.OK(w, ImageUploadResponse{
			ImageURL: imageURL,
			Message:  "Image uploaded successfully",
		})
	}
}

// GeneratePresignedUploadURL generates a presigned URL for direct upload
func GeneratePresignedUploadURL(s store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("home-id is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
			return
		}

		// Get content type from query params
		contentType := r.URL.Query().Get("content_type")
		if contentType == "" {
			contentType = "image/jpeg" // Default
		}

		if err := imageutil.ValidateImageType(contentType); err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Generate unique key for the image
		imageKey := imageutil.GenerateImageKey(userID, homeID, contentType)

		// Generate presigned URL
		uploadURL, err := s.GeneratePresignedURL(r.Context(), imageKey, contentType)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to generate presigned URL: %w", err))
			return
		}

		rest.OK(w, PresignedUploadResponse{
			UploadURL: uploadURL,
			ImageKey:  imageKey,
			Message:   "Presigned URL generated successfully",
		})
	}
}

// DeleteHomeImage handles deletion of home images
func DeleteHomeImage(s store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("home-id is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
			return
		}

		// Get image URL from request body
		var req struct {
			ImageURL string `json:"image_url"`
		}

		reqPayload, err := rest.ExtractPayload[struct {
			ImageURL string `json:"image_url"`
		}](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}
		req = reqPayload

		if req.ImageURL == "" {
			rest.ErrValidation(w, errors.New("image_url is required"))
			return
		}

		// Extract key from URL (this assumes S3Store implementation)
		if s3Store, ok := s.(*store.S3Store); ok {
			key, err := s3Store.ExtractKeyFromURL(req.ImageURL)
			if err != nil {
				rest.ErrBadRequest(w, fmt.Errorf("invalid image URL: %w", err))
				return
			}

			// Delete from storage
			if err := s.Delete(r.Context(), key); err != nil {
				rest.ErrInternal(w, fmt.Errorf("failed to delete image: %w", err))
				return
			}
		}

		rest.OK(w, map[string]string{
			"message": "Image deleted successfully",
		})
	}
}
