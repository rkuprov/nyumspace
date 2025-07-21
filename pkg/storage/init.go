package storage

import (
	"context"
	"slices"

	"github.com/rkuprov/nyumspace/pkg/config"
)

const (
	bucketImages = "images"
	bucketDocs   = "docs"
)

func init() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	if cfg.S3Aws == nil {
		panic("S3 configuration is not set")
	}
	ctx := context.Background()
	client, err := NewStorageClient(ctx, cfg.S3Aws)
	if err != nil {
		panic("Failed to create storage client: " + err.Error())
	}

	existing, err := client.listBuckets(ctx)
	if err != nil {
		panic("Failed to list buckets: " + err.Error())
	}
	for _, bucket := range []string{bucketImages, bucketDocs} {
		if slices.Contains(existing, bucket) {
			continue
		}
		if err := client.createBucket(ctx, bucket); err != nil {
			panic("Failed to create bucket " + bucket + ": " + err.Error())
		}
	}
}
