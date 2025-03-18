package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ssig33/fuckbase/internal/config"
	"github.com/ssig33/fuckbase/internal/logger"
)

// Client represents an S3 client
type Client struct {
	client     *minio.Client
	config     *config.S3Config
	bucketName string
}

// NewClient creates a new S3 client with the given configuration
func NewClient(cfg *config.S3Config) (*Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("S3 is not enabled in configuration")
	}

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false, // Set to true for HTTPS
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	client := &Client{
		client:     minioClient,
		config:     cfg,
		bucketName: cfg.Bucket,
	}

	// Ensure bucket exists
	if err := client.ensureBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return client, nil
}

// ensureBucketExists checks if the bucket exists and creates it if it doesn't
func (c *Client) ensureBucketExists() error {
	ctx := context.Background()
	exists, err := c.client.BucketExists(ctx, c.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		logger.Info("Creating bucket: %s", c.bucketName)
		err = c.client.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{
			Region: c.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to S3
func (c *Client) UploadFile(objectName string, data []byte, contentType string) error {
	ctx := context.Background()
	reader := bytes.NewReader(data)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("Successfully uploaded %s to %s", objectName, c.bucketName)
	return nil
}

// DownloadFile downloads a file from S3
func (c *Client) DownloadFile(objectName string) ([]byte, error) {
	ctx := context.Background()
	obj, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	logger.Info("Successfully downloaded %s from %s", objectName, c.bucketName)
	return data, nil
}

// ListFiles lists all files in the bucket with the given prefix
func (c *Client) ListFiles(prefix string) ([]string, error) {
	ctx := context.Background()
	objectCh := c.client.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objects []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// DeleteFile deletes a file from S3
func (c *Client) DeleteFile(objectName string) error {
	ctx := context.Background()
	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	logger.Info("Successfully deleted %s from %s", objectName, c.bucketName)
	return nil
}

// GetFileInfo gets information about a file in S3
func (c *Client) GetFileInfo(objectName string) (*minio.ObjectInfo, error) {
	ctx := context.Background()
	info, err := c.client.StatObject(ctx, c.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &info, nil
}

// GenerateBackupObjectName generates a unique object name for a backup
func GenerateBackupObjectName(databaseName string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("backups/%s/%s.json", databaseName, timestamp)
}

// GenerateFullBackupObjectName generates a unique object name for a full backup
func GenerateFullBackupObjectName() string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("backups/full/%s.json", timestamp)
}