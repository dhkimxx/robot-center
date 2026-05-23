package recording

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage interface {
	UploadManifest(ctx context.Context, objectKey string, manifest map[string]any) (int64, error)
	UploadFile(ctx context.Context, objectKey string, filePath string, contentType string) (int64, error)
}

type MinIOObjectStorage struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	mu        sync.Mutex
	client    *minio.Client
}

func NewMinIOObjectStorage(endpoint string, accessKey string, secretKey string, bucket string) *MinIOObjectStorage {
	return &MinIOObjectStorage{
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
	}
}

func (s *MinIOObjectStorage) UploadManifest(ctx context.Context, objectKey string, manifest map[string]any) (int64, error) {
	client, err := s.getClient()
	if err != nil {
		return 0, err
	}
	if err := s.ensureBucket(ctx, client); err != nil {
		return 0, err
	}

	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(payload)
	_, err = client.PutObject(ctx, s.bucket, objectKey, reader, int64(reader.Len()), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	return int64(reader.Len()), err
}

func (s *MinIOObjectStorage) UploadFile(ctx context.Context, objectKey string, filePath string, contentType string) (int64, error) {
	client, err := s.getClient()
	if err != nil {
		return 0, err
	}
	if err := s.ensureBucket(ctx, client); err != nil {
		return 0, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	_, err = client.PutObject(ctx, s.bucket, objectKey, file, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return stat.Size(), err
}

func (s *MinIOObjectStorage) ensureBucket(ctx context.Context, client *minio.Client) error {
	exists, err := client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return s.ensureBucketDownloadPolicy(ctx, client)
}

func (s *MinIOObjectStorage) ensureBucketDownloadPolicy(ctx context.Context, client *minio.Client) error {
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, s.bucket)
	return client.SetBucketPolicy(ctx, s.bucket, policy)
}

func (s *MinIOObjectStorage) getClient() (*minio.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		return s.client, nil
	}

	endpoint, secure, err := parseMinIOEndpoint(s.endpoint)
	if err != nil {
		return nil, err
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.accessKey, s.secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}
	s.client = client
	return client, nil
}

func parseMinIOEndpoint(rawEndpoint string) (string, bool, error) {
	if !strings.Contains(rawEndpoint, "://") {
		return rawEndpoint, false, nil
	}
	parsed, err := url.Parse(rawEndpoint)
	if err != nil {
		return "", false, err
	}
	if parsed.Host == "" {
		return "", false, fmt.Errorf("invalid MinIO endpoint %q", rawEndpoint)
	}
	return parsed.Host, parsed.Scheme == "https", nil
}
