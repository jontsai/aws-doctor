package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type service struct {
	client *s3.Client
}

// S3WasteInfo contains information about potential S3 waste
type S3WasteInfo struct {
	BucketName             string
	HasLifecyclePolicy     bool
	IncompleteMultipartCount int
	IncompleteMultipartSize  int64 // bytes
}

// S3Service defines the interface for S3 waste detection
type S3Service interface {
	GetS3WasteInfo(ctx context.Context) ([]S3WasteInfo, error)
}
