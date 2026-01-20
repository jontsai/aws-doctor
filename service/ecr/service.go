package ecr

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func NewService(awsconfig aws.Config) *service {
	client := ecr.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

// GetStaleImages returns ECR images that are potentially stale (older than staleDays)
func (s *service) GetStaleImages(ctx context.Context, staleDays int) ([]ECRRepositorySummary, error) {
	var summaries []ECRRepositorySummary

	// First, list all repositories
	repoInput := &ecr.DescribeRepositoriesInput{}
	repoPaginator := ecr.NewDescribeRepositoriesPaginator(s.client, repoInput)

	for repoPaginator.HasMorePages() {
		repoPage, err := repoPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, repo := range repoPage.Repositories {
			summary, err := s.analyzeRepository(ctx, repo, staleDays)
			if err != nil {
				// Log error but continue with other repos
				continue
			}

			// Only include repos that have waste
			if summary.StaleImages > 0 || summary.UntaggedImages > 0 {
				summaries = append(summaries, summary)
			}
		}
	}

	return summaries, nil
}

// analyzeRepository analyzes a single repository for stale/untagged images
func (s *service) analyzeRepository(ctx context.Context, repo types.Repository, staleDays int) (ECRRepositorySummary, error) {
	repoName := aws.ToString(repo.RepositoryName)
	summary := ECRRepositorySummary{
		RepositoryName: repoName,
	}

	cutoffTime := time.Now().AddDate(0, 0, -staleDays)

	// List all images in the repository
	imageInput := &ecr.DescribeImagesInput{
		RepositoryName: repo.RepositoryName,
	}
	imagePaginator := ecr.NewDescribeImagesPaginator(s.client, imageInput)

	for imagePaginator.HasMorePages() {
		imagePage, err := imagePaginator.NextPage(ctx)
		if err != nil {
			return summary, err
		}

		for _, imageDetail := range imagePage.ImageDetails {
			summary.TotalImages++

			sizeBytes := int64(0)
			if imageDetail.ImageSizeInBytes != nil {
				sizeBytes = *imageDetail.ImageSizeInBytes
			}
			summary.TotalSizeBytes += sizeBytes

			pushedAt := time.Time{}
			if imageDetail.ImagePushedAt != nil {
				pushedAt = *imageDetail.ImagePushedAt
			}

			daysSincePush := int(time.Since(pushedAt).Hours() / 24)
			isStale := pushedAt.Before(cutoffTime)
			isUntagged := len(imageDetail.ImageTags) == 0

			if isStale {
				summary.StaleImages++
				summary.StaleSizeBytes += sizeBytes
			}

			if isUntagged {
				summary.UntaggedImages++
			}

			// Only track images that are stale or untagged
			if isStale || isUntagged {
				// ECR pricing: $0.10 per GB per month
				estimatedCost := float64(sizeBytes) / (1024 * 1024 * 1024) * 0.10

				wasteInfo := ECRImageWasteInfo{
					RepositoryName: repoName,
					ImageDigest:    aws.ToString(imageDetail.ImageDigest),
					ImageTags:      imageDetail.ImageTags,
					PushedAt:       pushedAt.Format("2006-01-02"),
					SizeBytes:      sizeBytes,
					DaysSincePush:  daysSincePush,
					IsUntagged:     isUntagged,
					EstimatedCost:  estimatedCost,
				}

				summary.Images = append(summary.Images, wasteInfo)
				summary.EstimatedWaste += estimatedCost
			}
		}
	}

	return summary, nil
}
