package cloudwatchlogs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func NewService(awsconfig aws.Config) *service {
	client := cloudwatchlogs.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

// GetLogGroupsWithoutRetention returns log groups that have no retention policy set
// (infinite retention = indefinite storage costs)
func (s *service) GetLogGroupsWithoutRetention(ctx context.Context) ([]LogGroupWasteInfo, error) {
	var wasteInfo []LogGroupWasteInfo

	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, lg := range page.LogGroups {
			// Only include log groups without retention (infinite retention)
			if lg.RetentionInDays == nil {
				storedBytes := int64(0)
				if lg.StoredBytes != nil {
					storedBytes = *lg.StoredBytes
				}

				// CloudWatch Logs pricing: $0.03 per GB stored per month
				estimatedCost := float64(storedBytes) / (1024 * 1024 * 1024) * 0.03

				wasteInfo = append(wasteInfo, LogGroupWasteInfo{
					LogGroupName:  aws.ToString(lg.LogGroupName),
					RetentionDays: lg.RetentionInDays,
					StoredBytes:   storedBytes,
					EstimatedCost: estimatedCost,
				})
			}
		}
	}

	return wasteInfo, nil
}
