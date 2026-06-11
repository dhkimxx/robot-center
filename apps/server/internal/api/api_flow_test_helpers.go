package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/testsupport/postgrestest"
)

type apiFlowTestServer struct {
	baseURL string
}

type apiFlowTestServerOptions struct {
	recorderRuntimeBlockingReason string
	recorderRuntimeClearable      bool
}

type testRobot struct {
	code  string
	token string
}

type testMission struct {
	code string
	id   string
}

func newAPIFlowTestServer(t *testing.T) apiFlowTestServer {
	return newAPIFlowTestServerWithOptions(t, apiFlowTestServerOptions{
		recorderRuntimeClearable: true,
	})
}

func newAPIFlowTestServerWithOptions(t *testing.T, options apiFlowTestServerOptions) apiFlowTestServer {
	t.Helper()

	postgresContainer := postgrestest.Start(t)
	recorderHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/runtime/recordings/status" {
			recorderRuntime := map[string]any{
				"status":               "ok",
				"recordingDirectories": 2,
				"files":                4,
				"usedBytes":            4096,
				"totalBytes":           8192,
				"availableBytes":       4096,
				"usedPercent":          50,
				"clearable":            options.recorderRuntimeClearable,
			}
			if !options.recorderRuntimeClearable {
				recorderRuntime["blockingReason"] = options.recorderRuntimeBlockingReason
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"recorderRuntime": recorderRuntime,
			})
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/runtime/recordings/clear" {
			writeJSON(w, http.StatusOK, map[string]any{
				"recorderRuntime": map[string]any{
					"recordingDirectoriesDeleted": 2,
					"filesDeleted":                4,
					"deletedBytes":                4096,
				},
			})
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/runtime/recordings/prune" {
			writeJSON(w, http.StatusOK, map[string]any{
				"recorderRuntime": map[string]any{
					"recordingDirectoriesDeleted": 1,
					"filesDeleted":                2,
					"deletedBytes":                2048,
				},
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	}))
	t.Cleanup(recorderHealth.Close)

	appServer, err := NewServerFromConfig(context.Background(), config.AppServerConfig{
		PostgresDSN:               postgresContainer.DSN,
		AppServerPublicURL:        "http://center.local",
		RecorderWorkerInternalURL: recorderHealth.URL,
		SFUWebSocketPublicBaseURL: "ws://center.local",
		TURNPublicURL:             "turn:127.0.0.1:3478?transport=udp",
		TURNInternalURL:           "turn:127.0.0.1:3478?transport=udp",
		TURNUsername:              "robot",
		TURNPassword:              "robot-pass",
		MinIOInternalURL:          "http://127.0.0.1:9000",
		MinIOBucket:               "robot-center-poc",
	})
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(appServer.Handler())
	t.Cleanup(server.Close)
	return apiFlowTestServer{baseURL: server.URL}
}

func (s apiFlowTestServer) createRobot(t *testing.T, displayName string) testRobot {
	t.Helper()

	payload := requestJSON[dto.CreateRobotResponse](t, s.baseURL, http.MethodPost, "/api/v1/operator/robots", "", dto.CreateRobotRequest{
		DisplayName: displayName,
		ModelName:   "Android Mock",
	})
	return testRobot{
		code:  payload.Robot.RobotCode,
		token: payload.ConnectionInfo.RobotToken,
	}
}

func (s apiFlowTestServer) createMission(t *testing.T, robotCodes []string) testMission {
	t.Helper()

	payload := requestJSON[dto.MissionEnvelopeResponse](t, s.baseURL, http.MethodPost, "/api/v1/operator/missions", "", dto.CreateMissionRequest{
		Name:        "Integration Mission",
		MissionType: "mountain_rescue",
		SiteNote:    "test",
		RobotCodes:  robotCodes,
	})
	return testMission{
		code: payload.Mission.MissionCode,
		id:   payload.Mission.ID,
	}
}

func (s apiFlowTestServer) startMission(t *testing.T, missionCode string) dto.MissionResponse {
	t.Helper()

	payload := requestJSON[dto.MissionEnvelopeResponse](t, s.baseURL, http.MethodPost, "/api/v1/operator/missions/"+missionCode+"/start", "", nil)
	return payload.Mission
}

func (s apiFlowTestServer) createStartedMission(t *testing.T, robots ...testRobot) testMission {
	t.Helper()

	robotCodes := make([]string, 0, len(robots))
	for _, robot := range robots {
		robotCodes = append(robotCodes, robot.code)
	}
	mission := s.createMission(t, robotCodes)
	startedMission := s.startMission(t, mission.code)
	return testMission{
		code: startedMission.MissionCode,
		id:   startedMission.ID,
	}
}
