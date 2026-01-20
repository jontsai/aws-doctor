package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func NewService(awsconfig aws.Config) *service {
	client := s3.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

func (s *service) GetS3WasteInfo(ctx context.Context) ([]S3WasteInfo, error) {
	// List all buckets
	bucketsOutput, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var wasteInfo []S3WasteInfo

	for _, bucket := range bucketsOutput.Buckets {
		if bucket.Name == nil {
			continue
		}

		info := S3WasteInfo{
			BucketName: *bucket.Name,
		}

		// Check for lifecycle policy
		info.HasLifecyclePolicy = s.hasLifecyclePolicy(ctx, *bucket.Name)

		// Check for incomplete multipart uploads
		mpCount, mpSize := s.getIncompleteMultipartInfo(ctx, *bucket.Name)
		info.IncompleteMultipartCount = mpCount
		info.IncompleteMultipartSize = mpSize

		// Only add to waste if there's something to report
		if !info.HasLifecyclePolicy || info.IncompleteMultipartCount > 0 {
			wasteInfo = append(wasteInfo, info)
		}
	}

	return wasteInfo, nil
}

func (s *service) hasLifecyclePolicy(ctx context.Context, bucketName string) bool {
	_, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		// NoSuchLifecycleConfiguration means no policy exists
		var noSuchConfig *types.NoSuchLifecycleConfiguration
		if errors.As(err, &noSuchConfig) {
			return false
		}
		// Other errors (access denied, etc.) - assume policy might exist
		return true
	}

	return true
}

func (s *service) getIncompleteMultipartInfo(ctx context.Context, bucketName string) (int, int64) {
	var count int
	var totalSize int64

	input := &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucketName),
	}

	paginator := s3.NewListMultipartUploadsPaginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			// Ignore errors (access denied, etc.)
			break
		}

		for _, upload := range page.Uploads {
			count++

			// Get parts to calculate size
			partsInput := &s3.ListPartsInput{
				Bucket:   aws.String(bucketName),
				Key:      upload.Key,
				UploadId: upload.UploadId,
			}

			partsOutput, err := s.client.ListParts(ctx, partsInput)
			if err != nil {
				continue
			}

			for _, part := range partsOutput.Parts {
				if part.Size != nil {
					totalSize += *part.Size
				}
			}
		}
	}

	return count, totalSize
}
