package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
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
