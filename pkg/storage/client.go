package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/rkuprov/nyumspace/pkg/config"
)

type Client struct {
	client *s3.Client
}

// NewStorageClient creates a new S3-compatible storage client.
func NewStorageClient(ctx context.Context, cfg *config.S3Aws) (*Client, error) {
	s3conf, err := s3cfg.LoadDefaultConfig(ctx,
		s3cfg.WithRegion(cfg.Region),
		s3cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(s3conf, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(cfg.BaseEndpoint)
	})

	return &Client{client: client}, nil
}

// CreateBucket creates a new S3 bucket
func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// ListBuckets lists all buckets
func (c *Client) ListBuckets(ctx context.Context) ([]string, error) {
	result, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var buckets []string
	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	return buckets, nil
}

// PutObject uploads an object to S3
func (c *Client) PutObject(ctx context.Context, bucketName, key string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

// GetObject downloads an object from S3
func (c *Client) GetObject(ctx context.Context, bucketName, key string) ([]byte, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// DeleteObject deletes an object from S3
func (c *Client) DeleteObject(ctx context.Context, bucketName, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// ListObjects lists objects in a bucket
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := c.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return result.Contents, nil
}

// HeadObject gets object metadata
func (c *Client) HeadObject(ctx context.Context, bucketName, key string) (*s3.HeadObjectOutput, error) {
	result, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to head object: %w", err)
	}
	return result, nil
}

// CopyObject copies an object within the same bucket or between buckets
func (c *Client) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	copySource := fmt.Sprintf("%s/%s", srcBucket, srcKey)

	_, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}
	return nil
}

// ObjectExists checks if an object exists
func (c *Client) ObjectExists(ctx context.Context, bucketName, key string) (bool, error) {
	_, err := c.HeadObject(ctx, bucketName, key)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GeneratePresignedURL generates a presigned URL for object access
func (c *Client) GeneratePresignedURL(ctx context.Context, bucketName, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// UploadLargeFile uploads a large file using multipart upload
func (c *Client) UploadLargeFile(ctx context.Context, bucketName, key string, data []byte, partSize int64) error {
	// Create multipart upload
	createResp, err := c.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %w", err)
	}

	uploadID := createResp.UploadId
	var completedParts []types.CompletedPart

	// Upload parts
	dataLen := int64(len(data))
	partNumber := int32(1)

	for offset := int64(0); offset < dataLen; offset += partSize {
		end := offset + partSize
		if end > dataLen {
			end = dataLen
		}

		partData := data[offset:end]
		currentPart := partNumber

		uploadResp, err2 := c.client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(bucketName),
			Key:        aws.String(key),
			PartNumber: &currentPart,
			UploadId:   uploadID,
			Body:       bytes.NewReader(partData),
		})
		if err2 != nil {
			// Abort the multipart upload on error
			c.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucketName),
				Key:      aws.String(key),
				UploadId: uploadID,
			})
			return fmt.Errorf("failed to upload part %d: %w", partNumber, err2)
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       uploadResp.ETag,
			PartNumber: &currentPart,
		})

		log.Printf("Uploaded part %d: %s", currentPart, *uploadResp.ETag)
		partNumber++
	}

	log.Println("Uploading completed parts. Upload ID:", *uploadID)
	// Complete the multipart upload
	_, err = c.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}
