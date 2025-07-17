package store

import (
	"fmt"
)

// NewStore creates a new store instance based on the configuration
func NewStore(cfg *Config) (Store, error) {
	switch cfg.Provider {
	case "s3", "localstack":
		return NewS3Store(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}

// DefaultLocalStackConfig returns a default configuration for LocalStack
func DefaultLocalStackConfig(bucket string) *Config {
	return &Config{
		Provider:    "localstack",
		Region:      "us-east-1",
		Bucket:      bucket,
		Endpoint:    "http://localhost:4566",
		AccessKeyID: "test",
		SecretKey:   "test",
		PublicURL:   "http://localhost:4566",
	}
}

// DefaultAWSConfig returns a default configuration for AWS S3
func DefaultAWSConfig(region, bucket string) *Config {
	return &Config{
		Provider: "s3",
		Region:   region,
		Bucket:   bucket,
	}
}
