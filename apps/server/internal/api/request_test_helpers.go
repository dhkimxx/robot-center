package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func requestJSON[T any](t *testing.T, baseURL string, method string, path string, bearerToken string, body any) T {
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
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("%s %s returned %s", method, path, response.Status)
	}

	var payload T
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func requestRawJSON(t *testing.T, baseURL string, method string, path string, bearerToken string, body any) (int, map[string]any) {
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
	defer response.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response.StatusCode, payload
}

func componentHasStatus(components []any, name string, status string) bool {
	for _, item := range components {
		component, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if component["name"] == name && component["status"] == status {
			return true
		}
	}
	return false
}

func robotListHasCode(robots []any, robotCode string) bool {
	for _, item := range robots {
		robot, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if robot["robotCode"] == robotCode {
			return true
		}
	}
	return false
}

func sensorListHasID(sensors []any, sensorID string) bool {
	for _, item := range sensors {
		sensor, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if sensor["sensorId"] == sensorID {
			return true
		}
	}
	return false
}

func assertStringListEqual(t *testing.T, value any, expected []string) {
	t.Helper()

	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected string list, got %#v", value)
	}
	if len(items) != len(expected) {
		t.Fatalf("expected %d strings, got %#v", len(expected), value)
	}
	for index, expectedValue := range expected {
		actualValue, ok := items[index].(string)
		if !ok || actualValue != expectedValue {
			t.Fatalf("expected strings %#v, got %#v", expected, value)
		}
	}
}

func fileHasAvailableURL(files []any, fileType string) bool {
	for _, item := range files {
		file, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if file["type"] == fileType && file["status"] == "available" {
			urlValue, _ := file["url"].(string)
			return strings.Contains(urlValue, "http://center.local:9000/robot-center-poc/")
		}
	}
	return false
}
