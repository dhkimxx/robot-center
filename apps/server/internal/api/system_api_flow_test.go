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

	legacySwaggerResponse, err := http.Get(server.baseURL + "/api/docs")
	if err != nil {
		t.Fatalf("request legacy Swagger UI redirect: %v", err)
	}
	defer legacySwaggerResponse.Body.Close()
	if legacySwaggerResponse.Request.URL.Path != "/swagger/index.html" || legacySwaggerResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected legacy Swagger URL to redirect to /swagger/index.html, got %s %s", legacySwaggerResponse.Status, legacySwaggerResponse.Request.URL.Path)
	}

	systemSwaggerResponse, err := http.Get(server.baseURL + "/api/v1/system/docs")
	if err != nil {
		t.Fatalf("request system Swagger UI redirect: %v", err)
	}
	defer systemSwaggerResponse.Body.Close()
	if systemSwaggerResponse.Request.URL.Path != "/swagger/index.html" || systemSwaggerResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected system Swagger URL to redirect to /swagger/index.html, got %s %s", systemSwaggerResponse.Status, systemSwaggerResponse.Request.URL.Path)
	}
}
