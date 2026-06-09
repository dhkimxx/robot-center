package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"robot-center/apps/server/internal/store"
)

const clearObjectStorageConfirmation = "CLEAR_OBJECT_STORAGE"

var (
	ErrSystemActionForbidden            = errors.New("system action is disabled in production")
	ErrSystemActionConfirmationRequired = errors.New("confirmation is required")
	ErrSystemActionConflict             = errors.New("system action is blocked by active state")
)

type ObjectStorageAdminConfig struct {
	Environment string
	Endpoint    string
	Bucket      string
	AccessKey   string
	SecretKey   string
}

type ObjectStorageClearResult struct {
	Bucket                   string `json:"bucket"`
	DeletedObjectCount       int    `json:"deletedObjectCount"`
	DeletedBytes             int64  `json:"deletedBytes"`
	StorageObjectRowsDeleted int64  `json:"storageObjectRowsDeleted"`
	RecordingChunksReset     int64  `json:"recordingChunksReset"`
}

type ObjectStorageUsageResult struct {
	Status             string  `json:"status"`
	Bucket             string  `json:"bucket"`
	ObjectCount        int     `json:"objectCount"`
	BucketUsedBytes    int64   `json:"bucketUsedBytes"`
	TotalBytes         int64   `json:"totalBytes"`
	UsedBytes          int64   `json:"usedBytes"`
	AvailableBytes     int64   `json:"availableBytes"`
	UsedPercent        float64 `json:"usedPercent"`
	DiskTotalBytes     int64   `json:"diskTotalBytes"`
	DiskUsedBytes      int64   `json:"diskUsedBytes"`
	DiskAvailableBytes int64   `json:"diskAvailableBytes"`
	DiskUsedPercent    float64 `json:"diskUsedPercent"`
}

type ObjectStorageAdminService struct {
	config             ObjectStorageAdminConfig
	metadataRepository store.StorageAdminRepository
}

func NewObjectStorageAdminService(config ObjectStorageAdminConfig, metadataRepository store.StorageAdminRepository) *ObjectStorageAdminService {
	if strings.TrimSpace(config.Bucket) == "" {
		config.Bucket = "robot-center"
	}
	if strings.TrimSpace(config.AccessKey) == "" {
		config.AccessKey = "minioadmin"
	}
	if strings.TrimSpace(config.SecretKey) == "" {
		config.SecretKey = "minioadmin"
	}
	return &ObjectStorageAdminService{
		config:             config,
		metadataRepository: metadataRepository,
	}
}

func (s *ObjectStorageAdminService) ClearObjectStorage(ctx context.Context, confirmation string) (ObjectStorageClearResult, error) {
	if s == nil {
		return ObjectStorageClearResult{}, errors.New("object storage admin service is not configured")
	}
	if strings.EqualFold(strings.TrimSpace(s.config.Environment), "production") {
		return ObjectStorageClearResult{}, ErrSystemActionForbidden
	}
	if strings.TrimSpace(confirmation) != clearObjectStorageConfirmation {
		return ObjectStorageClearResult{}, ErrSystemActionConfirmationRequired
	}

	result, err := s.deleteBucketObjects(ctx)
	if err != nil {
		return ObjectStorageClearResult{}, err
	}
	if s.metadataRepository != nil {
		resetResult, err := s.metadataRepository.ResetObjectStorageMetadata(ctx)
		if err != nil {
			return ObjectStorageClearResult{}, err
		}
		result.StorageObjectRowsDeleted = resetResult.StorageObjectRowsDeleted
		result.RecordingChunksReset = resetResult.RecordingChunksReset
	}
	return result, nil
}

func (s *ObjectStorageAdminService) GetObjectStorageUsage(ctx context.Context) (ObjectStorageUsageResult, error) {
	if s == nil {
		return ObjectStorageUsageResult{}, errors.New("object storage admin service is not configured")
	}
	result := ObjectStorageUsageResult{
		Status: "ok",
		Bucket: strings.TrimSpace(s.config.Bucket),
	}

	client, err := s.minioClient()
	if err != nil {
		return result, err
	}
	exists, err := client.BucketExists(ctx, result.Bucket)
	if err != nil {
		return result, err
	}
	if exists {
		for object := range client.ListObjects(ctx, result.Bucket, minio.ListObjectsOptions{Recursive: true}) {
			if object.Err != nil {
				return result, object.Err
			}
			result.ObjectCount++
			result.BucketUsedBytes += object.Size
		}
	}

	adminClient, err := s.minioAdminClient()
	if err != nil {
		return result, err
	}
	storageInfo, err := adminClient.StorageInfo(ctx)
	if err != nil {
		return result, err
	}
	for _, disk := range storageInfo.Disks {
		if disk.TotalSpace == 0 && disk.UsedSpace == 0 && disk.AvailableSpace == 0 {
			continue
		}
		result.DiskTotalBytes += int64(disk.TotalSpace)
		result.DiskUsedBytes += int64(disk.UsedSpace)
		result.DiskAvailableBytes += int64(disk.AvailableSpace)
	}
	applyObjectStorageCapacity(&result)
	return result, nil
}

func (s *ObjectStorageAdminService) deleteBucketObjects(ctx context.Context) (ObjectStorageClearResult, error) {
	bucket := strings.TrimSpace(s.config.Bucket)
	result := ObjectStorageClearResult{Bucket: bucket}
	client, err := s.minioClient()
	if err != nil {
		return result, err
	}
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, nil
	}

	objects := make([]minio.ObjectInfo, 0)
	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return result, object.Err
		}
		result.DeletedObjectCount++
		result.DeletedBytes += object.Size
		objects = append(objects, minio.ObjectInfo{Key: object.Key})
	}

	objectsToRemove := make(chan minio.ObjectInfo, len(objects))
	for _, object := range objects {
		objectsToRemove <- object
	}
	close(objectsToRemove)
	for removeError := range client.RemoveObjects(ctx, bucket, objectsToRemove, minio.RemoveObjectsOptions{}) {
		if removeError.Err != nil {
			return result, removeError.Err
		}
	}
	return result, nil
}

func (s *ObjectStorageAdminService) minioClient() (*minio.Client, error) {
	endpoint, secure, err := parseStorageEndpoint(s.config.Endpoint)
	if err != nil {
		return nil, err
	}
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.config.AccessKey, s.config.SecretKey, ""),
		Secure: secure,
	})
}

func (s *ObjectStorageAdminService) minioAdminClient() (*madmin.AdminClient, error) {
	endpoint, secure, err := parseStorageEndpoint(s.config.Endpoint)
	if err != nil {
		return nil, err
	}
	return madmin.New(endpoint, s.config.AccessKey, s.config.SecretKey, secure)
}

func calculateStorageUsedPercent(usedBytes int64, totalBytes int64) float64 {
	if totalBytes <= 0 || usedBytes <= 0 {
		return 0
	}
	return float64(usedBytes) / float64(totalBytes) * 100
}

func applyObjectStorageCapacity(result *ObjectStorageUsageResult) {
	result.UsedBytes = result.BucketUsedBytes
	result.AvailableBytes = result.DiskAvailableBytes
	result.TotalBytes = result.UsedBytes + result.AvailableBytes
	result.UsedPercent = calculateStorageUsedPercent(result.UsedBytes, result.TotalBytes)
	result.DiskUsedPercent = calculateStorageUsedPercent(result.DiskUsedBytes, result.DiskTotalBytes)
}

func parseStorageEndpoint(rawEndpoint string) (string, bool, error) {
	rawEndpoint = strings.TrimSpace(rawEndpoint)
	if rawEndpoint == "" {
		return "", false, fmt.Errorf("MinIO endpoint is required")
	}
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
