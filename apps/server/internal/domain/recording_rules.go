package domain

import (
	"fmt"
	"time"
)

const DefaultRecordingChunkDurationSeconds = 600

type RecordingChunkWindow struct {
	Index           int
	StartedAt       time.Time
	EndedAt         time.Time
	DurationSeconds int
}

func NewRecordingChunkWindow(base time.Time, tickAt time.Time, durationSeconds int) RecordingChunkWindow {
	if durationSeconds <= 0 {
		durationSeconds = DefaultRecordingChunkDurationSeconds
	}
	if base.IsZero() {
		base = tickAt
	}

	elapsedSeconds := int(tickAt.Sub(base).Seconds())
	if elapsedSeconds < 0 {
		elapsedSeconds = 0
	}
	chunkIndex := elapsedSeconds / durationSeconds
	chunkStartedAt := base.Add(time.Duration(chunkIndex*durationSeconds) * time.Second)

	return RecordingChunkWindow{
		Index:           chunkIndex,
		StartedAt:       chunkStartedAt,
		EndedAt:         chunkStartedAt.Add(time.Duration(durationSeconds) * time.Second),
		DurationSeconds: durationSeconds,
	}
}

func NewRecordingObjectKeys(missionCode string, robotCode string, startedAt time.Time, endedAt time.Time) map[string]string {
	datePath := startedAt.UTC().Format("2006/01/02")
	timeRange := startedAt.UTC().Format("20060102T150405Z") + "_" + endedAt.UTC().Format("20060102T150405Z")
	basePath := fmt.Sprintf("missions/%s/robots/%s/recordings/%s/%s", missionCode, robotCode, datePath, timeRange)
	return map[string]string{
		"manifest":  basePath + "_manifest.json",
		"rgbMp4":    basePath + "_rgb_h264_opus.mp4",
		"thermal":   basePath + "_thermal_h264.mp4",
		"sensor":    basePath + "_sensor.jsonl",
		"telemetry": basePath + "_telemetry.jsonl",
	}
}

func NewRecordingManifest(chunk RecordingChunk) map[string]any {
	return map[string]any{
		"schemaVersion":      "1.0",
		"chunkId":            chunk.ID,
		"recordingSessionId": chunk.RecordingSessionID,
		"missionId":          chunk.MissionID,
		"missionCode":        chunk.MissionCode,
		"robotCode":          chunk.RobotCode,
		"chunkIndex":         chunk.ChunkIndex,
		"status":             chunk.Status,
		"startedAt":          chunk.StartedAt.Format(time.RFC3339Nano),
		"endedAt":            chunk.EndedAt.Format(time.RFC3339Nano),
		"codecPolicy": map[string]string{
			"video": "h264",
			"audio": "opus",
		},
		"mediaObjectKeys":    chunk.MediaObjectKeys,
		"availableFileTypes": chunk.AvailableFileTypes,
	}
}
