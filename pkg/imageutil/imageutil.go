package imageutil

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SupportedImageTypes contains the MIME types for supported image formats
var SupportedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// ValidateImageType checks if the content type is a supported image type
func ValidateImageType(contentType string) error {
	if !SupportedImageTypes[contentType] {
		return fmt.Errorf("unsupported image type: %s", contentType)
	}
	return nil
}

// GenerateImageKey generates a unique key for storing an image
func GenerateImageKey(userID, homeID, contentType string) string {
	timestamp := time.Now().Unix()
	uniqueID := uuid.New().String()

	// Get file extension from content type
	ext := getExtensionFromContentType(contentType)

	return fmt.Sprintf("images/homes/%s/%s/%d_%s%s", userID, homeID, timestamp, uniqueID, ext)
}

// GeneratePresignedUploadKey generates a key for presigned uploads
func GeneratePresignedUploadKey(userID, homeID string) string {
	timestamp := time.Now().Unix()
	uniqueID := uuid.New().String()

	return fmt.Sprintf("images/homes/%s/%s/%d_%s", userID, homeID, timestamp, uniqueID)
}

// getExtensionFromContentType returns the file extension for a content type
func getExtensionFromContentType(contentType string) string {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		// Fallback to common extensions
		switch contentType {
		case "image/jpeg", "image/jpg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/gif":
			return ".gif"
		case "image/webp":
			return ".webp"
		default:
			return ".jpg" // Default fallback
		}
	}
	return extensions[0]
}

// ExtractContentTypeFromFilename extracts content type from filename
func ExtractContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
