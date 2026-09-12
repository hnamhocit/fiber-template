package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/hnamhocit/fiber-template/internal/config"
)

// Client embeds *minio.Client plus the default bucket from config.
type Client struct {
	*minio.Client
	bucket string
}

// New builds the MinIO client, or (nil, nil) when not configured.
func New(svc config.Services) (*Client, error) {
	if !svc.MinioEnabled() {
		return nil, nil
	}
	mc, err := minio.New(svc.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(svc.MinioUser, svc.MinioPassword, ""),
		Secure: svc.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Client{Client: mc, bucket: svc.MinioBucket}, nil
}

// Bucket returns the configured default bucket name.
func (c *Client) Bucket() string { return c.bucket }

// EnsureBucket creates the default bucket if it does not exist.
// Idempotent: safe to call on every boot, no-op when the bucket is there.
func (c *Client) EnsureBucket(ctx context.Context) error {
	if c.bucket == "" {
		return fmt.Errorf("MINIO_BUCKET is not set")
	}
	exists, err := c.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", c.bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %q: %w", c.bucket, err)
	}
	return nil
}

// HealthCheck verifies connectivity AND that the configured bucket exists.
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.bucket == "" {
		return fmt.Errorf("MINIO_BUCKET is not set")
	}
	ok, err := c.BucketExists(ctx, c.bucket)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("bucket %q not found", c.bucket)
	}
	return nil
}

// Available reports whether object storage is configured.
// Safe to call on a nil *Client.
func (c *Client) Available() bool { return c != nil }
