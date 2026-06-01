package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestSystemAPIFlow(t *testing.T) {
	server := newAPIFlowTestServer(t)

	health := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/healthz", "", nil)
	if health["status"] != "ok" {
		t.Fatalf("expected health ok, got %#v", health)
	}

	systemStatus := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/api/v1/system/status", "", nil)
	components := systemStatus["components"].([]any)
	if !componentHasStatus(components, "recorder-worker", "ok") {
		t.Fatalf("expected recorder-worker component status ok, got %#v", components)
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
