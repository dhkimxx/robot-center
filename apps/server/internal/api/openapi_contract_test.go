package api

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"robot-center/apps/server/internal/api/dto"
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
	if len(paths) != len(expectedPaths) {
		t.Fatalf("expected %d OpenAPI paths, got %d: %#v", len(expectedPaths), len(paths), paths)
	}
	for _, expectedPath := range expectedPaths {
		if _, ok := paths[expectedPath]; !ok {
			t.Fatalf("expected OpenAPI path %s, got %#v", expectedPath, paths)
		}
	}
	for path := range paths {
		if path != "/healthz" && !strings.HasPrefix(path, "/api/v1/") {
			t.Fatalf("expected role-based /api/v1 path or /healthz, got %s", path)
		}
		if strings.Contains(path, "/api/v1/recoder/") {
			t.Fatalf("recoder typo must not be documented, got %s", path)
		}
	}
}

func TestOpenAPISchemasMatchDTOFields(t *testing.T) {
	server := newAPIFlowTestServer(t)
	openAPI := requestJSON[map[string]any](t, server.baseURL, http.MethodGet, "/swagger/openapi.json", "", nil)

	type schemaContract struct {
		name string
		dto  any
	}
	contracts := []schemaContract{
		{name: "HealthResponse", dto: dto.HealthResponse{}},
		{name: "SystemStatusResponse", dto: dto.SystemStatusResponse{}},
		{name: "SystemComponentStatus", dto: dto.SystemComponentStatus{}},
		{name: "SystemConfig", dto: dto.SystemConfigResponse{}},
		{name: "ObjectStorageStatus", dto: dto.ObjectStorageStatusResponse{}},
		{name: "SystemSummary", dto: dto.SystemSummaryResponse{}},
		{name: "RTCConfigResponse", dto: dto.RTCConfigResponse{}},
		{name: "RecordingTargetsResponse", dto: dto.RecordingTargetsResponse{}},
		{name: "RecordingsResponse", dto: dto.RecordingsResponse{}},
		{name: "RecordingChunkEnvelope", dto: dto.RecordingChunkEnvelopeResponse{}},
		{name: "RecordingChunk", dto: dto.RecordingChunkResponse{}},
		{name: "RecordingFile", dto: dto.RecordingFileResponse{}},
		{name: "RecordingTickResponse", dto: dto.RecordingTickResponse{}},
		{name: "RecorderUploadRequest", dto: dto.RecorderUploadRequest{}},
		{name: "RecorderFinalizationClaimRequest", dto: dto.RecorderFinalizationClaimRequest{}},
		{name: "RecorderFinalizationStatusRequest", dto: dto.RecorderFinalizationStatusRequest{}},
		{name: "RecorderFinalizationJobsResponse", dto: dto.RecorderFinalizationJobsResponse{}},
		{name: "RecorderFinalizationJob", dto: dto.RecordingFinalizationJobResponse{}},
		{name: "OKResponse", dto: dto.OKResponse{}},
		{name: "RobotsResponse", dto: dto.RobotsResponse{}},
		{name: "RobotEnvelope", dto: dto.RobotEnvelopeResponse{}},
		{name: "CreateRobotRequest", dto: dto.CreateRobotRequest{}},
		{name: "UpdateRobotRequest", dto: dto.UpdateRobotRequest{}},
		{name: "CreateRobotResponse", dto: dto.CreateRobotResponse{}},
		{name: "RobotConnectionInfoEnvelope", dto: dto.RobotConnectionInfoEnvelopeResponse{}},
		{name: "Robot", dto: dto.RobotResponse{}},
		{name: "RobotConnectionInfo", dto: dto.RobotConnectionInfoResponse{}},
		{name: "MissionsResponse", dto: dto.MissionsResponse{}},
		{name: "MissionResponseEnvelope", dto: dto.MissionEnvelopeResponse{}},
		{name: "Mission", dto: dto.MissionResponse{}},
		{name: "CreateMissionRequest", dto: dto.CreateMissionRequest{}},
		{name: "MissionLiveStatusResponse", dto: dto.MissionLiveStatusResponse{}},
		{name: "RobotHeartbeatRequest", dto: dto.RobotHeartbeatRequest{}},
		{name: "RobotHeartbeatResponse", dto: dto.RobotHeartbeatResponse{}},
		{name: "RobotMissionResponse", dto: dto.RobotMissionResponse{}},
		{name: "RobotSFUConfig", dto: dto.RobotSFUConfigResponse{}},
		{name: "TurnServer", dto: dto.RobotTurnServerResponse{}},
		{name: "ErrorResponse", dto: dto.ErrorResponse{}},
	}
	for _, contract := range contracts {
		assertOpenAPISchemaMatchesDTOFields(t, openAPI, contract.name, contract.dto)
	}
}

func assertOpenAPISchemaMatchesDTOFields(t *testing.T, openAPI map[string]any, schemaName string, dtoValue any) {
	t.Helper()

	schemas := openAPI["components"].(map[string]any)["schemas"].(map[string]any)
	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI schema %s", schemaName)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI schema %s properties, got %#v", schemaName, schema)
	}

	expected := dtoJSONFields(dtoValue)
	actual := make(map[string]struct{}, len(properties))
	for field := range properties {
		actual[field] = struct{}{}
	}
	for field := range expected {
		if _, ok := actual[field]; !ok {
			t.Fatalf("expected OpenAPI schema %s to include DTO field %q; schema=%#v", schemaName, field, properties)
		}
	}
	for field := range actual {
		if _, ok := expected[field]; !ok {
			t.Fatalf("expected OpenAPI schema %s not to include non-DTO field %q; dtoFields=%#v schema=%#v", schemaName, field, expected, properties)
		}
	}
}

func dtoJSONFields(dtoValue any) map[string]struct{} {
	dtoType := reflect.TypeOf(dtoValue)
	for dtoType.Kind() == reflect.Pointer {
		dtoType = dtoType.Elem()
	}

	fields := map[string]struct{}{}
	for index := 0; index < dtoType.NumField(); index++ {
		field := dtoType.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if field.Anonymous {
			for embeddedField := range dtoJSONFields(reflect.New(field.Type).Elem().Interface()) {
				fields[embeddedField] = struct{}{}
			}
			continue
		}
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}
		fields[jsonName] = struct{}{}
	}
	return fields
}
