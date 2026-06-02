package dto

import (
	"encoding/json"
	"errors"
	"testing"

	"robot-center/apps/server/internal/service"
)

func TestObjectStorageStatusIncludesZeroUsageFields(t *testing.T) {
	response := ObjectStorageStatus(service.ObjectStorageUsageResult{
		Status: "ok",
		Bucket: "robot-center-poc",
	})

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	for _, field := range []string{
		"objectCount",
		"bucketUsedBytes",
		"totalBytes",
		"usedBytes",
		"availableBytes",
		"usedPercent",
		"diskTotalBytes",
		"diskUsedBytes",
		"diskAvailableBytes",
		"diskUsedPercent",
	} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("expected field %q in object storage status JSON, got %s", field, string(payload))
		}
	}
}

func TestObjectStorageUnavailableOmitsUsageFields(t *testing.T) {
	response := ObjectStorageUnavailable("robot-center-poc", errors.New("minio unavailable"))

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if fields["status"] != "unavailable" || fields["bucket"] != "robot-center-poc" || fields["error"] != "minio unavailable" {
		t.Fatalf("unexpected unavailable object storage response: %#v", fields)
	}
	if _, ok := fields["objectCount"]; ok {
		t.Fatalf("expected usage fields to be omitted for unavailable object storage, got %s", string(payload))
	}
}
