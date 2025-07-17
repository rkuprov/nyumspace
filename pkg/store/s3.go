package store

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Store implements the Store interface using AWS S3 or LocalStack
type S3Store struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewS3Store creates a new S3Store instance
func NewS3Store(cfg *Config) (*S3Store, error) {
	var awsCfg aws.Config
	var err error

	if cfg.Endpoint != "" {
		// Configure for LocalStack
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		})

		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretKey,
				"",
			)),
		)
	} else {
		// Configure for AWS S3
		if cfg.AccessKeyID != "" && cfg.SecretKey != "" {
			awsCfg, err = config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(cfg.Region),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					cfg.AccessKeyID,
					cfg.SecretKey,
					"",
				)),
			)
		} else {
			// Use default credential chain
			awsCfg, err = config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(cfg.Region),
			)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			// Force path-style addressing for LocalStack
			o.UsePathStyle = true
		}
	})

	return &S3Store{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

// Upload uploads a file to S3 and returns the URL
func (s *S3Store) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Make the object publicly readable
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return s.GetURL(ctx, key)
}

// Delete deletes a file from S3
func (s *S3Store) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

// GetURL returns the public URL for a given key
func (s *S3Store) GetURL(ctx context.Context, key string) (string, error) {
	if s.publicURL != "" {
		// Use custom public URL (e.g., for LocalStack)
		baseURL, err := url.Parse(s.publicURL)
		if err != nil {
			return "", fmt.Errorf("invalid public URL: %w", err)
		}
		baseURL.Path = fmt.Sprintf("/%s/%s", s.bucket, key)
		return baseURL.String(), nil
	}

	// Use standard AWS S3 URL format
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key), nil
}

// GeneratePresignedURL generates a presigned URL for direct upload
func (s *S3Store) GeneratePresignedURL(ctx context.Context, key string, contentType string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	}

	presignedURL, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 15 // URL expires in 15 minutes
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.URL, nil
}

// ExtractKeyFromURL extracts the S3 key from a URL
func (s *S3Store) ExtractKeyFromURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Handle different URL formats
	if s.publicURL != "" {
		// Custom public URL format
		baseURL, err := url.Parse(s.publicURL)
		if err != nil {
			return "", fmt.Errorf("invalid public URL: %w", err)
		}

		if parsedURL.Host == baseURL.Host {
			// Remove bucket name from path
			path := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", s.bucket))
			return path, nil
		}
	}

	// Standard AWS S3 URL format
	if strings.Contains(parsedURL.Host, ".s3.") {
		return strings.TrimPrefix(parsedURL.Path, "/"), nil
	}

	return "", fmt.Errorf("unable to extract key from URL: %s", urlStr)
}
