package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSFUWebSocketRoutesWorkThroughGin(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "WebSocket Robot")
	mission := server.createStartedMission(t, robot)

	websocketURL := "ws" + strings.TrimPrefix(server.baseURL, "http")
	operator := dialAPIWebSocket(t, websocketURL+"/api/v1/operator/sfu/ws?room="+mission.code, "")
	defer operator.Close()
	assertAPIWebSocketMessageType(t, operator, "joined")

	recorder := dialAPIWebSocket(t, websocketURL+"/api/v1/recorder/sfu/ws?room="+mission.code, "")
	defer recorder.Close()
	assertAPIWebSocketMessageType(t, recorder, "joined")

	robotPublisher := dialAPIWebSocket(t, websocketURL+"/api/v1/robot/sfu/ws?room="+mission.code, robot.token)
	defer robotPublisher.Close()
	assertAPIWebSocketMessageType(t, robotPublisher, "joined")
}

func TestSFUWebSocketRoutesRejectInvalidRoleParametersThroughGin(t *testing.T) {
	server := newAPIFlowTestServer(t)
	robot := server.createRobot(t, "WebSocket Reject Robot")
	mission := server.createStartedMission(t, robot)

	websocketURL := "ws" + strings.TrimPrefix(server.baseURL, "http")
	assertAPIWebSocketStatus(t, websocketURL+"/sfu/ws?room="+mission.code+"&role=robot", "", http.StatusNotFound)
	assertAPIWebSocketStatus(t, websocketURL+"/api/v1/operator/sfu/ws?room="+mission.code+"&robotCode="+robot.code, "", http.StatusBadRequest)
	assertAPIWebSocketStatus(t, websocketURL+"/api/v1/recorder/sfu/ws?room="+mission.code+"&robotCode="+robot.code, "", http.StatusBadRequest)
	assertAPIWebSocketStatus(t, websocketURL+"/api/v1/robot/sfu/ws?room="+mission.code, "", http.StatusUnauthorized)
	assertAPIWebSocketStatus(t, websocketURL+"/api/v1/robot/sfu/ws?room="+mission.code+"&role=robot", robot.token, http.StatusBadRequest)
}

func dialAPIWebSocket(t *testing.T, url string, bearerToken string) *websocket.Conn {
	t.Helper()

	headers := websocketHeaders(bearerToken)
	conn, response, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		t.Fatalf("expected websocket dial to succeed, url=%s status=%s err=%v", url, responseStatus(response), err)
	}
	return conn
}

func assertAPIWebSocketStatus(t *testing.T, url string, bearerToken string, expectedStatus int) {
	t.Helper()

	headers := websocketHeaders(bearerToken)
	conn, response, err := websocket.DefaultDialer.Dial(url, headers)
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Fatalf("expected websocket dial to fail with %d, url=%s", expectedStatus, url)
	}
	if response == nil || response.StatusCode != expectedStatus {
		t.Fatalf("expected websocket status %d, got response=%s err=%v", expectedStatus, responseStatus(response), err)
	}
}

func assertAPIWebSocketMessageType(t *testing.T, conn *websocket.Conn, expectedType string) {
	t.Helper()

	_, rawMessage, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var message struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rawMessage, &message); err != nil {
		t.Fatalf("decode websocket message: %v", err)
	}
	if message.Type != expectedType {
		t.Fatalf("expected websocket message type %q, got %q raw=%s", expectedType, message.Type, string(rawMessage))
	}
}

func websocketHeaders(bearerToken string) http.Header {
	headers := http.Header{}
	if strings.TrimSpace(bearerToken) != "" {
		headers.Set("Authorization", "Bearer "+bearerToken)
	}
	return headers
}

func responseStatus(response *http.Response) string {
	if response == nil {
		return "<nil>"
	}
	return response.Status
}
