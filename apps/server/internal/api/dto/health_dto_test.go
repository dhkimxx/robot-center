package dto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHealthResponseShape(t *testing.T) {
	startedAt := time.Date(2026, 6, 2, 7, 0, 1, 123456789, time.UTC)

	payload, err := json.Marshal(Health(startedAt))
	if err != nil {
		t.Fatalf("marshal health response: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}
	if fields["status"] != "ok" || fields["service"] != "app-server" {
		t.Fatalf("unexpected health response: %#v", fields)
	}
	if fields["startedAt"] != "2026-06-02T07:00:01Z" {
		t.Fatalf("startedAt = %q, want RFC3339 second precision", fields["startedAt"])
	}
}
