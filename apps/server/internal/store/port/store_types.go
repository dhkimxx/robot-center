package port

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidState = errors.New("invalid state")
)

type MissionStartConflict struct {
	RobotCode         string
	ActiveMissionCode string
}

type MissionStartConflictError struct {
	Conflicts []MissionStartConflict
}

func (e *MissionStartConflictError) Error() string {
	if e == nil || len(e.Conflicts) == 0 {
		return ErrInvalidState.Error()
	}
	parts := make([]string, 0, len(e.Conflicts))
	for _, conflict := range e.Conflicts {
		parts = append(parts, fmt.Sprintf("%s already active in %s", conflict.RobotCode, conflict.ActiveMissionCode))
	}
	return "mission start conflict: " + strings.Join(parts, ", ")
}

func (e *MissionStartConflictError) Unwrap() error {
	return ErrInvalidState
}

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

type RecordingSession struct {
	ID        string
	StartedAt time.Time
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
	WorkerID  string
	Attempt   int
}
