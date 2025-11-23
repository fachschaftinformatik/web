package buckets

import (
	"context"
	"fmt"
	"io"

	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	minioClient *minio.Client
	bucket      string
	domain   string
}

func NewClient(cfg *config.Config) (*Client, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		minioClient: client,
		bucket:      cfg.S3Bucket,
		domain:   cfg.Domain,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = c.minioClient.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := c.minioClient.PutObject(ctx, c.bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (c *Client) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	return c.minioClient.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
}

func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.minioClient.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}
