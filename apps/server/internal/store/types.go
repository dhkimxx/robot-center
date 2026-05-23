package store

import (
	"errors"
	"time"

	"robot-center/apps/server/internal/domain"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidState = errors.New("invalid state")
)

type CreateRobotInput struct {
	DisplayName string
	ModelName   string
}

type UpdateRobotInput struct {
	DisplayName string
	ModelName   string
}

type CreateMissionInput struct {
	Name        string
	MissionType string
	SiteNote    string
	RobotCode   string
	RobotCodes  []string
}

type HeartbeatInput struct {
	RobotCode      string
	State          string
	BatteryPercent int
	NetworkQuality string
	SentAt         time.Time
}

type RecordingTickInput struct {
	MissionCode          string
	RobotCode            string
	ChunkDurationSeconds int
	TickAt               time.Time
}

type RecordingTarget struct {
	Mission   domain.Mission
	RobotID   string
	RobotCode string
}

type CreateRecordingChunkInput struct {
	RecordingSessionID string
	MissionID          string
	MissionCode        string
	RobotID            string
	RobotCode          string
	Window             domain.RecordingChunkWindow
	MediaObjectKeys    map[string]string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type RecordingUploadMetadata struct {
	SizeBytes *int64
	Checksum  string
}
