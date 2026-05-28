package service

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestObjectStorageClearRequiresConfirmation(t *testing.T) {
	service := NewObjectStorageAdminService(ObjectStorageAdminConfig{
		Environment: "development",
		Endpoint:    "http://127.0.0.1:9000",
	}, nil)

	_, err := service.ClearObjectStorage(context.Background(), "wrong")
	if !errors.Is(err, ErrSystemActionConfirmationRequired) {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestObjectStorageClearIsDisabledInProduction(t *testing.T) {
	service := NewObjectStorageAdminService(ObjectStorageAdminConfig{
		Environment: "production",
		Endpoint:    "http://127.0.0.1:9000",
	}, nil)

	_, err := service.ClearObjectStorage(context.Background(), clearObjectStorageConfirmation)
	if !errors.Is(err, ErrSystemActionForbidden) {
		t.Fatalf("expected production guard error, got %v", err)
	}
}

func TestCalculateStorageUsedPercent(t *testing.T) {
	percent := calculateStorageUsedPercent(25, 100)
	if percent != 25 {
		t.Fatalf("expected 25 percent, got %f", percent)
	}

	percent = calculateStorageUsedPercent(25, 0)
	if percent != 0 {
		t.Fatalf("expected 0 percent for zero capacity, got %f", percent)
	}
}

func TestObjectStorageUsageUsesBucketObjectsAsUsedBytes(t *testing.T) {
	result := ObjectStorageUsageResult{
		BucketUsedBytes:    10,
		DiskAvailableBytes: 90,
		DiskTotalBytes:     200,
		DiskUsedBytes:      110,
	}

	applyObjectStorageCapacity(&result)

	if result.UsedBytes != 10 {
		t.Fatalf("expected object storage used bytes to follow bucket objects, got %d", result.UsedBytes)
	}
	if result.TotalBytes != 100 {
		t.Fatalf("expected object storage total to be bucket used plus available bytes, got %d", result.TotalBytes)
	}
	if math.Abs(result.UsedPercent-10) > 0.0001 {
		t.Fatalf("expected object storage used percent to be 10, got %f", result.UsedPercent)
	}
	if math.Abs(result.DiskUsedPercent-55) > 0.0001 {
		t.Fatalf("expected disk used percent to remain separately available, got %f", result.DiskUsedPercent)
	}
}
