package api

import (
	"net/http"
	"strings"
	"testing"

	"robot-center/apps/server/internal/api/dto"
)

func TestSystemAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)

	health := requestJSON[dto.HealthResponse](t, server.baseURL, http.MethodGet, "/healthz", "", nil)
	if health.Status != "ok" {
		t.Fatalf("expected health ok, got %#v", health)
	}

	systemStatus := requestJSON[dto.SystemStatusResponse](t, server.baseURL, http.MethodGet, "/api/v1/system/status", "", nil)
	if !componentHasStatus(systemStatus.Components, "recorder-worker", "ok") {
		t.Fatalf("expected recorder-worker component status ok, got %#v", systemStatus.Components)
	}
	if systemStatus.Database.Status != "ok" || systemStatus.Database.DatabaseSizeBytes == nil {
		t.Fatalf("expected database usage status, got %#v", systemStatus.Database)
	}
	if systemStatus.RecorderRuntime.Status != "ok" || systemStatus.RecorderRuntime.UsedBytes != 4096 || !systemStatus.RecorderRuntime.Clearable {
		t.Fatalf("expected recorder runtime status, got %#v", systemStatus.RecorderRuntime)
	}

	swaggerResponse, err := http.Get(server.baseURL + "/swagger/index.html")
	if err != nil {
		t.Fatalf("request Swagger UI: %v", err)
	}
	defer swaggerResponse.Body.Close()
	if swaggerResponse.StatusCode != http.StatusOK || !strings.Contains(swaggerResponse.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("expected Swagger UI HTML response, got %s %s", swaggerResponse.Status, swaggerResponse.Header.Get("Content-Type"))
	}
}

func TestClearObjectStorageIsBlockedWhenRecorderRuntimeIsActive(t *testing.T) {
	server := newAPIFlowTestServerWithOptions(t, apiFlowTestServerOptions{
		recorderRuntimeBlockingReason: "active recording target",
		recorderRuntimeClearable:      false,
	})

	status := requestStatus(t, server.baseURL, http.MethodPost, "/api/v1/system/object-storage/clear", "", dto.ClearObjectStorageRequest{
		Confirmation: "CLEAR_OBJECT_STORAGE",
	})
	if status != http.StatusConflict {
		t.Fatalf("expected object storage clear to be blocked by active recorder runtime, got %d", status)
	}
}

func TestSystemPruneActionsRunWhenRecorderRuntimeIsActive(t *testing.T) {
	server := newAPIFlowTestServerWithOptions(t, apiFlowTestServerOptions{
		recorderRuntimeBlockingReason: "active recording target",
		recorderRuntimeClearable:      false,
	})

	objectStoragePayload := requestJSON[dto.PruneObjectStorageResponse](t, server.baseURL, http.MethodPost, "/api/v1/system/object-storage/prune", "", dto.PruneObjectStorageRequest{
		Confirmation: "PRUNE_OBJECT_STORAGE",
	})
	if objectStoragePayload.ObjectStorage.DeletedObjectCount != 0 {
		t.Fatalf("expected empty object storage prune on fresh test database, got %#v", objectStoragePayload)
	}

	recorderRuntimePayload := requestJSON[dto.PruneRecorderRuntimeResponse](t, server.baseURL, http.MethodPost, "/api/v1/system/recorder-runtime/prune", "", dto.PruneRecorderRuntimeRequest{
		Confirmation: "PRUNE_RECORDER_RUNTIME",
	})
	if recorderRuntimePayload.RecorderRuntime.FilesDeleted != 2 || recorderRuntimePayload.RecorderRuntime.DeletedBytes != 2048 {
		t.Fatalf("expected recorder runtime prune result, got %#v", recorderRuntimePayload)
	}
}
