package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/testsupport/postgrestest"
)

type apiFlowTestServer struct {
	baseURL string
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
	t.Helper()

	postgresContainer := postgrestest.Start(t)
	recorderHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	}))
	t.Cleanup(recorderHealth.Close)

	appServer, err := NewServerFromConfig(context.Background(), config.AppServerConfig{
		PostgresDSN:               postgresContainer.DSN,
		PublicURL:                 "http://center.local",
		RecorderWorkerURL:         recorderHealth.URL,
		SFUWebSocketPublicBaseURL: "ws://center.local",
		TURNPublicURL:             "turn:127.0.0.1:3478?transport=udp",
		TURNInternalURL:           "turn:127.0.0.1:3478?transport=udp",
		TURNUsername:              "robot",
		TURNPassword:              "robot-pass",
		MinIOEndpoint:             "http://127.0.0.1:9000",
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

	payload := requestJSON[map[string]any](t, s.baseURL, http.MethodPost, "/api/v1/operator/robots", "", map[string]any{
		"displayName": displayName,
		"modelName":   "Android Mock",
	})
	robot := payload["robot"].(map[string]any)
	connectionInfo := payload["connectionInfo"].(map[string]any)
	return testRobot{
		code:  robot["robotCode"].(string),
		token: connectionInfo["robotToken"].(string),
	}
}

func (s apiFlowTestServer) createMission(t *testing.T, robotCodes []string) testMission {
	t.Helper()

	payload := requestJSON[map[string]any](t, s.baseURL, http.MethodPost, "/api/v1/operator/missions", "", map[string]any{
		"name":        "Integration Mission",
		"missionType": "mountain_rescue",
		"siteNote":    "test",
		"robotCodes":  robotCodes,
	})
	mission := payload["mission"].(map[string]any)
	return testMission{
		code: mission["missionCode"].(string),
		id:   mission["id"].(string),
	}
}

func (s apiFlowTestServer) startMission(t *testing.T, missionCode string) map[string]any {
	t.Helper()

	payload := requestJSON[map[string]any](t, s.baseURL, http.MethodPost, "/api/v1/operator/missions/"+missionCode+"/start", "", nil)
	return payload["mission"].(map[string]any)
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
		code: startedMission["missionCode"].(string),
		id:   startedMission["id"].(string),
	}
}
