package cloudwatchlogs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type service struct {
	client *cloudwatchlogs.Client
}

// LogGroupWasteInfo contains information about potential CloudWatch Logs waste
type LogGroupWasteInfo struct {
	LogGroupName     string
	RetentionDays    *int32 // nil means infinite retention
	StoredBytes      int64
	EstimatedCost    float64 // Monthly cost estimate
}

// CloudWatchLogsService defines the interface for CloudWatch Logs waste detection
type CloudWatchLogsService interface {
	GetLogGroupsWithoutRetention(ctx context.Context) ([]LogGroupWasteInfo, error)
}
