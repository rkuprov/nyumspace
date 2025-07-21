package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxImageSize = 5 * 1024 * 1024 // 5MB
)

func (c *Client) AddImage(ctx context.Context, userID, filename string, payload io.Reader) (string, error) {
	bucketName := bucketImages
	key := fmt.Sprintf("%s/%d_%s", userID, time.Now().UnixMilli(), filename)
	var contentType string
	switch filepath.Ext(filename) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case "":
		return "", fmt.Errorf("filename cannot be empty")
	default:
		return "", fmt.Errorf("unsupported image format: %s", filepath.Ext(filename))
	}

	//Read up to maxImageSize + 1 bytes to check size
	contents, err := io.ReadAll(io.LimitReader(payload, maxImageSize+1))
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	switch {
	case len(contents) == 0:
		return "", fmt.Errorf("image data cannot be empty")
	case len(contents) > maxImageSize:
		return "", fmt.Errorf("file size is limited to no greater than %d MiB", maxImageSize/(1024*1024))
	}

	_, err = c.client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(key),
			Body:          bytes.NewReader(contents),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(int64(len(contents))),
		})
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	return key, nil
}

// GetImage retrieves an image from S3 by userID and key. It returns the image as an io.ReadCloser to avoid loading the
// entire image into memory at once. The caller is responsible for closing the returned io.ReadCloser.
func (c *Client) GetImage(ctx context.Context, userID, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if !strings.HasPrefix(key, userID) {
		return nil, fmt.Errorf("key %s does not belong to user %s", key, userID)
	}

	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketImages),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return result.Body, nil
}

func (c *Client) DeleteImage(ctx context.Context, userID, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if !strings.HasPrefix(key, userID) {
		return fmt.Errorf("key %s does not belong to user %s", key, userID)
	}

	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketImages),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
