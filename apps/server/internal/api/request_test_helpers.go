package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"robot-center/apps/server/internal/api/dto"
)

func requestJSON[T any](t *testing.T, baseURL string, method string, path string, bearerToken string, body any) T {
	t.Helper()

	response := sendJSONRequest(t, baseURL, method, path, bearerToken, body)
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("%s %s returned %s", method, path, response.Status)
	}

	var payload T
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func requestRawJSONAs[T any](t *testing.T, baseURL string, method string, path string, bearerToken string, body any) (int, T) {
	t.Helper()

	response := sendJSONRequest(t, baseURL, method, path, bearerToken, body)
	defer response.Body.Close()

	var payload T
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response.StatusCode, payload
}

func requestStatus(t *testing.T, baseURL string, method string, path string, bearerToken string, body any) int {
	t.Helper()

	response := sendJSONRequest(t, baseURL, method, path, bearerToken, body)
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	return response.StatusCode
}

func sendJSONRequest(t *testing.T, baseURL string, method string, path string, bearerToken string, body any) *http.Response {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		rawBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(rawBody)
	}

	request, err := http.NewRequest(method, baseURL+path, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(bearerToken) != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("send request %s %s: %v", method, path, err)
	}
	return response
}

func componentHasStatus(components []dto.SystemComponentStatus, name string, status string) bool {
	for _, component := range components {
		if component.Name == name && component.Status == status {
			return true
		}
	}
	return false
}

func robotListHasCode(robots []dto.RobotResponse, robotCode string) bool {
	for _, robot := range robots {
		if robot.RobotCode == robotCode {
			return true
		}
	}
	return false
}

func sensorListHasID(sensors []dto.SensorLatestResponse, sensorID string) bool {
	for _, sensor := range sensors {
		if sensor.SensorID == sensorID {
			return true
		}
	}
	return false
}

func assertStringListEqual(t *testing.T, actual []string, expected []string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("expected strings %#v, got %#v", expected, actual)
	}
	for index, expectedValue := range expected {
		if actual[index] != expectedValue {
			t.Fatalf("expected strings %#v, got %#v", expected, actual)
		}
	}
}

func fileHasAvailableURL(files []dto.RecordingFileResponse, fileType string) bool {
	for _, file := range files {
		if file.Type == fileType && file.Status == "available" {
			return strings.Contains(file.URL, "http://center.local:9000/robot-center-poc/")
		}
	}
	return false
}
