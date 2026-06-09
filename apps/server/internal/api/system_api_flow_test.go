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
