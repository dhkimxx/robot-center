package api

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestSwaggerDocsUseRoleBasedPathsAndRobotSecurity(t *testing.T) {
	server := newAPIFlowTestServer(t)
	swaggerDoc := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/swagger/doc.json", "", nil)
	openAPIAlias := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/swagger/openapi.json", "", nil)

	assertGeneratedSwaggerContract(t, "Swagger", swaggerDoc)
	assertGeneratedSwaggerContract(t, "OpenAPI alias", openAPIAlias)
	if !reflect.DeepEqual(swaggerDoc["paths"], openAPIAlias["paths"]) {
		t.Fatalf("expected /swagger/openapi.json to expose the generated Swagger paths")
	}
	if !reflect.DeepEqual(swaggerDoc["securityDefinitions"], openAPIAlias["securityDefinitions"]) {
		t.Fatalf("expected /swagger/openapi.json to expose the generated Swagger security definitions")
	}
}

func assertGeneratedSwaggerContract(t *testing.T, label string, swaggerDoc map[string]any) {
	t.Helper()

	if swaggerDoc["swagger"] != "2.0" {
		t.Fatalf("expected %s to serve Swagger 2.0 doc, got %#v", label, swaggerDoc["swagger"])
	}
	info := swaggerDoc["info"].(map[string]any)
	if !strings.Contains(info["description"].(string), "관제 서버") {
		t.Fatalf("expected Korean %s description, got %#v", label, info)
	}
	forbiddenOpenAPIWords := []string{"개발 서버", "테스트 슬롯", "테스트용", "Mock Robot"}
	for _, forbiddenWord := range forbiddenOpenAPIWords {
		if strings.Contains(info["description"].(string), forbiddenWord) {
			t.Fatalf("%s description should be API-reference oriented, got %#v", label, info)
		}
	}

	paths := swaggerDoc["paths"].(map[string]any)
	assertRoleBasedAPIPaths(t, label, paths)

	securityDefinitions := swaggerDoc["securityDefinitions"].(map[string]any)
	robotSecurity, ok := securityDefinitions["RobotBearerAuth"].(map[string]any)
	if !ok {
		t.Fatalf("expected %s RobotBearerAuth security definition, got %#v", label, securityDefinitions)
	}
	if robotSecurity["in"] != "header" || robotSecurity["name"] != "Authorization" {
		t.Fatalf("expected %s RobotBearerAuth to use Authorization header, got %#v", label, robotSecurity)
	}

	robotHeartbeatPath := paths["/api/v1/robot/heartbeat"].(map[string]any)
	robotHeartbeatPost := robotHeartbeatPath["post"].(map[string]any)
	security := robotHeartbeatPost["security"].([]any)
	if len(security) != 1 {
		t.Fatalf("expected %s robot heartbeat to require RobotBearerAuth, got %#v", label, security)
	}
}

func assertRoleBasedAPIPaths(t *testing.T, label string, paths map[string]any) {
	t.Helper()

	expectedPaths := expectedRoleBasedAPIPaths()
	if len(paths) != len(expectedPaths) {
		t.Fatalf("expected %d %s paths, got %d: %#v", len(expectedPaths), label, len(paths), paths)
	}
	for _, expectedPath := range expectedPaths {
		if _, ok := paths[expectedPath]; !ok {
			t.Fatalf("expected %s path %s, got %#v", label, expectedPath, paths)
		}
	}
	for path := range paths {
		if path != "/healthz" && !strings.HasPrefix(path, "/api/v1/") {
			t.Fatalf("expected role-based /api/v1 path or /healthz in %s, got %s", label, path)
		}
		if strings.Contains(path, "/api/v1/recoder/") {
			t.Fatalf("recoder typo must not be documented in %s, got %s", label, path)
		}
	}
}

func expectedRoleBasedAPIPaths() []string {
	return []string{
		"/healthz",
		"/api/v1/system/status",
		"/api/v1/system/object-storage/clear",
		"/api/v1/system/sensors/clear",
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
}
