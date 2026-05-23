//go:build integration

package storage

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rkuprov/nyumspace/scratch/pkg/app/config"
)

func TestS3Client(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	require.NoError(t, err)
	client, err := NewStorageClient(ctx, cfg.S3Aws)
	require.NoError(t, err)

	bucketName := uuid.NewString()
	log.Println("Creating a new bucket...")
	err = client.CreateBucket(ctx, bucketName)
	assert.NoError(t, err, "Failed to create bucket")
	log.Println("Created bucket:", bucketName)

	log.Println("Creating a new object...")
	testData := []byte("Hello, LocalStack S3 Interface!")

	err = client.PutObject(ctx, bucketName, "test/hello.txt", testData, "text/plain")
	assert.NoError(t, err, "Failed to create object")

	jsonData := []byte(`{"message": "Hello from SeaweedFS", "timestamp": "2024-01-01T00:00:00Z"}`)
	err = client.PutObject(ctx, bucketName, "data/sample.json", jsonData, "application/json")
	assert.NoError(t, err, "Failed to upload json object")

	log.Println("Listing buckets...")
	buckets, err := client.ListBuckets(ctx)
	assert.NoError(t, err, "Failed to list buckets")
	assert.Greater(t, len(buckets), 0, "No buckets found")

	// Example 4: List objects
	log.Println("Listing objects...")
	objects, err := client.ListObjects(ctx, bucketName, "")
	assert.NoError(t, err, "Failed to list objects")
	log.Printf("Found %d objects\n", len(objects))
	for _, obj := range objects {
		log.Printf("- %s (size: %d, modified: %s)\n", *obj.Key, *obj.Size, obj.LastModified.Format(time.RFC3339))
	}

	// Example 5: Download object
	log.Println("Downloading object...")
	downloadedData, err := client.GetObject(ctx, bucketName, "test/hello.txt")
	assert.NoError(t, err, "Failed to download object")
	log.Printf("Downloaded data: %s\n", string(downloadedData))

	// Example 6: Object metadata
	fmt.Println("\n=== Example 6: Object Metadata ===")
	metadata, err := client.HeadObject(ctx, bucketName, "test/hello.txt")
	if err != nil {
		log.Printf("Failed to get object metadata: %v", err)
	} else {
		fmt.Printf("Content-Type: %s\n", *metadata.ContentType)
		fmt.Printf("Content-Length: %d\n", *metadata.ContentLength)
		fmt.Printf("Last-Modified: %s\n", metadata.LastModified.Format(time.RFC3339))
	}

	// Example 7: Copy object
	fmt.Println("\n=== Example 7: Copy Object ===")
	err = client.CopyObject(ctx, bucketName, "test/hello.txt", bucketName, "test/hello_copy.txt")
	if err != nil {
		log.Printf("Failed to copy object: %v", err)
	} else {
		fmt.Println("Object copied successfully")
	}

	// Example 8: Check if object exists
	fmt.Println("\n=== Example 8: Check Object Existence ===")
	exists, err := client.ObjectExists(ctx, bucketName, "test/hello.txt")
	if err != nil {
		log.Printf("Failed to check object existence: %v", err)
	} else {
		fmt.Printf("Object exists: %v\n", exists)
	}

	// Example 9: Generate presigned URL
	fmt.Println("\n=== Example 9: Generate Presigned URL ===")
	presignedURL, err := client.GeneratePresignedURLGet(ctx, bucketName, "test/hello.txt", 1*time.Hour)
	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
	} else {
		fmt.Printf("Presigned URL: %s\n", presignedURL)
	}

	// Example 10: Large file upload (multipart)
	fmt.Println("\n=== Example 10: Multipart Upload ===")
	largeData := make([]byte, 10*1024*1024) // 10MB of data
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	err = client.UploadLargeFile(ctx, bucketName, "large/big_file.bin", largeData, 6*1024*1024)
	if err != nil {
		log.Printf("Failed to upload large file: %v", err)
	} else {
		fmt.Println("Large file uploaded successfully using multipart upload")
	}

	// Example 11: Cleanup - delete objects
	//fmt.Println("\n=== Example 11: Cleanup ===")
	//objectsToDelete := []string{"test/hello.txt", "test/hello_copy.txt", "data/sample.json", "large/big_file.bin"}

	//for _, key := range objectsToDelete {
	//	err = client.DeleteObject(ctx, bucketName, key)
	//	if err != nil {
	//		log.Printf("Failed to delete object %s: %v", key, err)
	//	} else {
	//		fmt.Printf("Deleted object: %s\n", key)
	//	}
	//}
}
