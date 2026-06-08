package monitorlog

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestEventFormatsKeyValueLine(t *testing.T) {
	var buffer bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	defer func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
	}()

	log.SetOutput(&buffer)
	log.SetFlags(0)

	Event("sfu", "robot_offer_received", "room", "mission-001", "reason", "contains spaces", "empty", "")

	got := strings.TrimSpace(buffer.String())
	want := `sfu monitor event=robot_offer_received room=mission-001 reason="contains spaces"`
	if got != want {
		t.Fatalf("unexpected monitor log line:\ngot  %q\nwant %q", got, want)
	}
}

func TestFormatValueTruncatesLongValues(t *testing.T) {
	value := strings.Repeat("a", maxLogValueLength+10)
	got := formatValue(value)
	if strings.Count(got, "a") != maxLogValueLength || !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated log value, got length=%d value=%q", len(got), got)
	}
}
