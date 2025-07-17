package store

import (
	"context"
	"io"
)

// Store defines the interface for storage operations
type Store interface {
	// Upload uploads a file to the storage and returns the URL
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)

	// Delete deletes a file from storage
	Delete(ctx context.Context, key string) error

	// GetURL returns the public URL for a given key
	GetURL(ctx context.Context, key string) (string, error)

	// GeneratePresignedURL generates a presigned URL for direct upload
	GeneratePresignedURL(ctx context.Context, key string, contentType string) (string, error)
}

// Config holds configuration for storage providers
type Config struct {
	Provider    string // "s3" or "localstack"
	Region      string
	Bucket      string
	Endpoint    string // For localstack
	AccessKeyID string
	SecretKey   string
	PublicURL   string // Base URL for public access
}
