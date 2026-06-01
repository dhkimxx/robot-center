package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestOpenAPIContractUsesRoleBasedPaths(t *testing.T) {
	server := newAPIFlowTestServer(t)
	openAPI := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/swagger/openapi.json", "", nil)

	info := openAPI["info"].(map[string]any)
	if !strings.Contains(info["description"].(string), "관제 서버") {
		t.Fatalf("expected Korean OpenAPI description, got %#v", info)
	}
	forbiddenOpenAPIWords := []string{"개발 서버", "테스트 슬롯", "테스트용", "Mock Robot"}
	for _, forbiddenWord := range forbiddenOpenAPIWords {
		if strings.Contains(info["description"].(string), forbiddenWord) {
			t.Fatalf("OpenAPI description should be API-reference oriented, got %#v", info)
		}
	}

	paths := openAPI["paths"].(map[string]any)
	expectedPaths := []string{
		"/healthz",
		"/swagger/index.html",
		"/swagger/openapi.json",
		"/api/v1/system/status",
		"/api/v1/system/object-storage/clear",
		"/api/v1/operator/rtc-config",
		"/api/v1/operator/sensor-descriptors",
		"/api/v1/operator/sensor-samples",
		"/api/v1/operator/sensor-latest",
		"/api/v1/operator/recordings",
		"/api/v1/operator/robots",
		"/api/v1/operator/robots/{robotCode}",
		"/api/v1/operator/robots/{robotCode}/connection-info",
		"/api/v1/operator/robots/{robotCode}/connection-token",
		"/api/v1/operator/missions",
		"/api/v1/operator/missions/{missionCode}/live-status",
		"/api/v1/operator/missions/{missionCode}/start",
		"/api/v1/operator/missions/{missionCode}/end",
		"/api/v1/operator/sfu/ws",
		"/api/v1/recorder/recording-targets",
		"/api/v1/recorder/tick",
		"/api/v1/recorder/finalization-jobs/claim",
		"/api/v1/recorder/finalization-jobs/{jobID}/completed",
		"/api/v1/recorder/finalization-jobs/{jobID}/partial",
		"/api/v1/recorder/finalization-jobs/{jobID}/failed",
		"/api/v1/recorder/chunks/{chunkID}/uploaded",
		"/api/v1/recorder/chunks/{chunkID}/files/{fileType}/uploaded",
		"/api/v1/recorder/sensor-samples",
		"/api/v1/recorder/sfu/ws",
		"/api/v1/robot/heartbeat",
		"/api/v1/robot/mission",
		"/api/v1/robot/sfu/ws",
	}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("expected %d OpenAPI paths, got %d: %#v", len(expectedPaths), len(paths), paths)
	}
	for _, expectedPath := range expectedPaths {
		if _, ok := paths[expectedPath]; !ok {
			t.Fatalf("expected OpenAPI path %s, got %#v", expectedPath, paths)
		}
	}
	for path := range paths {
		if path != "/healthz" && !strings.HasPrefix(path, "/api/v1/") && !strings.HasPrefix(path, "/swagger/") {
			t.Fatalf("expected role-based /api/v1 path, /swagger path, or /healthz, got %s", path)
		}
		if strings.Contains(path, "/api/v1/recoder/") {
			t.Fatalf("recoder typo must not be documented, got %s", path)
		}
	}
}
