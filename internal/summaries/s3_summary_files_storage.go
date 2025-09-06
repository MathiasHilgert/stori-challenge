package summaries

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3SummaryFilesStorage implements SummaryFilesStorage using AWS S3 as backend.
type S3SummaryFilesStorage struct {
	client *s3.Client
}

// NewS3SummaryFilesStorage creates a new instance backed by S3.
func NewS3SummaryFilesStorage(client *s3.Client) *S3SummaryFilesStorage {
	return &S3SummaryFilesStorage{client: client}
}

// Get retrieves a SummaryFile from S3 by path ("s3://bucket/key" or "bucket/key").
// It fetches both file content and metadata from object tags.
func (s *S3SummaryFilesStorage) Get(ctx context.Context, path string) (*SummaryFile, error) {
	bucket, key, err := s.parsePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 path %q: %w", path, err)
	}

	// Fetch content
	content, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object s3://%s/%s: %w", bucket, key, err)
	}

	// Fetch metadata (tags)
	accountID, accountEmail, err := s.getFileMetadata(ctx, bucket, key)
	if err != nil {
		return nil, err
	}

	return &SummaryFile{
		Path:         fmt.Sprintf("s3://%s/%s", bucket, key),
		AccountID:    accountID,
		AccountEmail: accountEmail,
		Content:      content.Body,
	}, nil
}

// parsePath extracts bucket and key from "s3://bucket/key" or "bucket/key".
func (s *S3SummaryFilesStorage) parsePath(path string) (bucket, key string, err error) {
	if path == "" {
		return "", "", fmt.Errorf("path is empty")
	}
	parts := strings.SplitN(strings.TrimPrefix(path, "s3://"), "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("path must be 'bucket/key' or 's3://bucket/key'")
	}
	return parts[0], parts[1], nil
}

// getFileMetadata extracts AccountID and AccountEmail from S3 object tags.
func (s *S3SummaryFilesStorage) getFileMetadata(ctx context.Context, bucket, key string) (string, string, error) {
	tags, err := s.client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", "", fmt.Errorf("get tags s3://%s/%s: %w", bucket, key, err)
	}

	var accountID, accountEmail string
	for _, t := range tags.TagSet {
		if t.Key == nil || t.Value == nil {
			continue
		}
		switch strings.ToLower(*t.Key) {
		case "accountid", "account_id":
			accountID = *t.Value
		case "accountemail", "account_email", "email":
			accountEmail = *t.Value
		}
	}

	if accountID == "" || accountEmail == "" {
		return "", "", fmt.Errorf("missing required tags in s3://%s/%s (found: AccountID=%q, AccountEmail=%q)", bucket, key, accountID, accountEmail)
	}
	return accountID, accountEmail, nil
}
