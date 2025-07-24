package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3/log"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
)

// GetHomeImage retrieves a home image by imageID
func GetHomeImage(h *homes.Homes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrValidation(w, errors.New("no user id provided"))
			return
		}

		imageID := chi.URLParam(r, "image-id")
		if imageID == "" {
			rest.ErrValidation(w, errors.New("no image id provided"))
			return
		}

		// Get image key from database
		var imageKey string
		var dbUserID string
		err := h.DB.QueryRow(r.Context(),
			"SELECT image_key, user_id FROM nyum_images WHERE id = $1",
			imageID).Scan(&imageKey, &dbUserID)
		if err != nil {
			rest.ErrNotFound(w, err)
			return
		}

		// Verify user owns this image
		if dbUserID != userID {
			rest.ErrUnauthorized(w, errors.New("not authorized"))
			log.Error(fmt.Errorf("not authorized. User ID: %s, Image ID: %s", userID, imageID))
			return
		}

		// Get image from S3
		imageReader, err := h.Store.GetImage(r.Context(), userID, imageKey)
		if err != nil {
			rest.ErrInternal(w, err)
			return
		}
		defer func() {
			if imageReader.Close() != nil {
				log.Error(fmt.Errorf("failed to close image. err: %w", err))
			}
		}()

		// Determine content type from image key extension
		ext := filepath.Ext(imageKey)
		var contentType string
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		default:
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year

		if _, err = io.Copy(w, imageReader); err != nil {
			// If we've already started writing, we can't send an error status
			log.Error(err)
			return
		}
	}
}

// UploadHomeImage handles image upload for a home
func UploadHomeImage(h *homes.Homes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrValidation(w, errors.New("no user id provided"))
			return
		}

		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("no home id provided"))
			return
		}

		// Verify user owns this home
		var ownerID string
		err := h.DB.QueryRow(r.Context(),
			"SELECT owner_id FROM homes WHERE id = $1",
			homeID).Scan(&ownerID)
		if err != nil {
			rest.ErrNotFound(w, fmt.Errorf("home not found: %w", err))
			return
		}

		if ownerID != userID {
			rest.ErrUnauthorized(w, errors.New("not authorized"))
			log.Error(fmt.Errorf("not authorized. User ID: %s. Owner id: %s", userID, ownerID))
			return
		}

		// Parse multipart form
		if err = r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
			rest.ErrInternal(w, err)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}
		defer func() {
			if file.Close() != nil {
				log.Error(fmt.Errorf("failed to close file. err: %w", err))
			}
		}()

		// Upload to S3
		imageKey, err := h.Store.AddImage(r.Context(), userID, header.Filename, file)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to upload image to S3: %w", err))
			return
		}

		// Save image record to database
		imageID := uuid.NewString()
		_, err = h.DB.Exec(r.Context(),
			"INSERT INTO nyum_images (id, user_id, image_key, home_id) VALUES ($1, $2, $3, $4)",
			imageID, userID, imageKey, homeID)
		if err != nil {
			// Try to clean up S3 object if database insert fails
			err2 := h.Store.DeleteImage(r.Context(), userID, imageKey)
			if err2 != nil {
				rest.ErrInternal(w, errors.Join(err2, err))
				log.Error(fmt.Errorf("failed to save image to db: %w", err2))
			} else {
				rest.ErrInternal(w, fmt.Errorf("failed to save image to db: %w", err))
				log.Error(fmt.Errorf("failed to save image record: %w", err))
			}
			return
		}

		rest.Created(
			w,
			nyum.ImageCreateResponse{
				ImageID: imageID,
				HomeID:  homeID,
			})
	}
}

// DeleteHomeImage deletes a home image by imageID
func DeleteHomeImage(h *homes.Homes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrValidation(w, errors.New("no user id provided"))
			return
		}

		imageID := chi.URLParam(r, "image-id")
		if imageID == "" {
			rest.ErrValidation(w, errors.New("no image id provided"))
			return
		}

		// Get image details from database
		var imageKey string
		var ownerID string
		var homeID string
		err := h.DB.QueryRow(r.Context(),
			"SELECT image_key, user_id, home_id FROM nyum_images WHERE id = $1",
			imageID).Scan(&imageKey, &ownerID, &homeID)
		if err != nil {
			rest.ErrNotFound(w, fmt.Errorf("image not found: %w", err))
			return
		}

		// Verify user owns this image
		if ownerID != userID {
			rest.ErrUnauthorized(w, errors.New("not authorized"))
			log.Error(fmt.Errorf("not authorized. User ID: %s, Image ID: %s, Owner ID: %s", userID, imageID, ownerID))
			return
		}

		// Delete from S3
		if err = h.Store.DeleteImage(r.Context(), userID, imageKey); err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to delete image from S3: %w", err))
			return
		}

		// Delete from database
		_, err = h.DB.Exec(r.Context(),
			"DELETE FROM nyum_images WHERE id = $1",
			imageID)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to delete image from DB: %w", err))
			return
		}
		// Return success response
		rest.ResultOK(
			w,
			rest.Result[any]{
				Message: "Image deleted successfully",
			},
		)
	}
}
