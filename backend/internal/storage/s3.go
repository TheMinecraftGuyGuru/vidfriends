package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/vidfriends/backend/internal/config"
)

// S3Storage implements videos.AssetStorage backed by an S3-compatible service.
type S3Storage struct {
	uploader *manager.Uploader
	bucket   string
	baseURL  string
}

// NewS3Storage configures an uploader targeting the provided object store.
func NewS3Storage(ctx context.Context, cfg config.ObjectStoreConfig) (*S3Storage, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("s3 storage: bucket is required")
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if strings.TrimSpace(cfg.Endpoint) != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:           cfg.Endpoint,
					SigningRegion: cfg.Region,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		loadOpts = append(loadOpts, awsconfig.WithEndpointResolverWithOptions(resolver))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = false
	})

	return &S3Storage{
		uploader: uploader,
		bucket:   cfg.Bucket,
		baseURL:  strings.TrimSuffix(cfg.PublicBaseURL, "/"),
	}, nil
}

// Save uploads the provided content to the configured bucket and returns a public location.
func (s *S3Storage) Save(ctx context.Context, name string, r io.Reader) (string, error) {
	key := strings.TrimLeft(name, "/")
	if key == "" {
		return "", fmt.Errorf("s3 storage: empty key")
	}

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   manager.ReadSeekCloser(r),
		ACL:    s3types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("s3 storage upload %s: %w", key, err)
	}

	if s.baseURL == "" {
		return key, nil
	}

	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}
