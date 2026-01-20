package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type service struct {
	client *ecr.Client
}

// ECRImageWasteInfo contains information about potentially stale ECR images
type ECRImageWasteInfo struct {
	RepositoryName string
	ImageDigest    string
	ImageTags      []string
	PushedAt       string    // When the image was pushed
	SizeBytes      int64     // Image size in bytes
	DaysSincePush  int       // Days since the image was pushed
	IsUntagged     bool      // Whether the image has no tags
	EstimatedCost  float64   // Monthly storage cost estimate
}

// ECRRepositorySummary summarizes waste in a repository
type ECRRepositorySummary struct {
	RepositoryName     string
	TotalImages        int
	StaleImages        int               // Images older than threshold
	UntaggedImages     int               // Images without tags
	TotalSizeBytes     int64
	StaleSizeBytes     int64
	EstimatedWaste     float64           // Monthly cost of stale images
	Images             []ECRImageWasteInfo
}

// ECRService defines the interface for ECR waste detection
type ECRService interface {
	GetStaleImages(ctx context.Context, staleDays int) ([]ECRRepositorySummary, error)
}
