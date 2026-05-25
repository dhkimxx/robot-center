package dto

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type RobotResponse struct {
	ID              string     `json:"id"`
	RobotCode       string     `json:"robotCode"`
	DisplayName     string     `json:"displayName"`
	ModelName       string     `json:"modelName,omitempty"`
	Status          string     `json:"status"`
	LastSeenAt      *time.Time `json:"lastSeenAt,omitempty"`
	LastStreamingAt *time.Time `json:"lastStreamingAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type RobotConnectionInfoResponse struct {
	ServerURL  string `json:"serverUrl"`
	RobotCode  string `json:"robotCode"`
	RobotToken string `json:"robotToken"`
}

type MissionResponse struct {
	ID          string     `json:"id"`
	MissionCode string     `json:"missionCode"`
	Name        string     `json:"name"`
	MissionType string     `json:"missionType"`
	Status      string     `json:"status"`
	SiteNote    string     `json:"siteNote,omitempty"`
	RobotCode   string     `json:"robotCode,omitempty"`
	RobotCodes  []string   `json:"robotCodes,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type StreamingTrackResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Kind        string `json:"kind"`
	Codec       string `json:"codec"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	FPS         int    `json:"fps,omitempty"`
	BitrateKbps int    `json:"bitrateKbps,omitempty"`
}

type StreamingStatusResponse struct {
	RobotCode             string                   `json:"robotCode"`
	MissionID             string                   `json:"missionId"`
	RoomID                string                   `json:"roomId"`
	Status                string                   `json:"status"`
	PublishedTracks       []StreamingTrackResponse `json:"publishedTracks"`
	PublishedDataChannels []string                 `json:"publishedDataChannels"`
	SentAt                time.Time                `json:"sentAt"`
}

type TelemetryResponse struct {
	ID                  string          `json:"id"`
	RobotCode           string          `json:"robotCode"`
	MissionID           string          `json:"missionId"`
	MessageID           string          `json:"messageId,omitempty"`
	MessageType         string          `json:"messageType,omitempty"`
	Sequence            int64           `json:"sequence,omitempty"`
	SentAt              *time.Time      `json:"sentAt,omitempty"`
	ReceivedAt          time.Time       `json:"receivedAt"`
	BatteryPercent      *float64        `json:"batteryPercent,omitempty"`
	NetworkState        string          `json:"networkState,omitempty"`
	PositionAvailable   bool            `json:"positionAvailable"`
	Latitude            *float64        `json:"latitude,omitempty"`
	Longitude           *float64        `json:"longitude,omitempty"`
	AltitudeMeter       *float64        `json:"altitudeMeter,omitempty"`
	AccuracyMeter       *float64        `json:"accuracyMeter,omitempty"`
	HeadingDegree       *float64        `json:"headingDegree,omitempty"`
	SpeedMeterPerSecond *float64        `json:"speedMeterPerSecond,omitempty"`
	RawPayload          json.RawMessage `json:"rawPayload"`
}

type SensorReadingResponse struct {
	ID                 string          `json:"id"`
	RobotCode          string          `json:"robotCode"`
	MissionID          string          `json:"missionId"`
	MessageID          string          `json:"messageId,omitempty"`
	MessageType        string          `json:"messageType,omitempty"`
	Sequence           int64           `json:"sequence,omitempty"`
	SentAt             *time.Time      `json:"sentAt,omitempty"`
	ReceivedAt         time.Time       `json:"receivedAt"`
	BatteryPercent     *float64        `json:"batteryPercent,omitempty"`
	TemperatureCelsius *float64        `json:"temperatureCelsius,omitempty"`
	HumidityPercent    *float64        `json:"humidityPercent,omitempty"`
	OxygenPercent      *float64        `json:"oxygenPercent,omitempty"`
	COPpm              *float64        `json:"coPpm,omitempty"`
	CH4Ppm             *float64        `json:"ch4Ppm,omitempty"`
	RawPayload         json.RawMessage `json:"rawPayload"`
}

type RecordingChunkResponse struct {
	ID                 string                  `json:"id"`
	RecordingSessionID string                  `json:"recordingSessionId"`
	MissionID          string                  `json:"missionId"`
	MissionCode        string                  `json:"missionCode"`
	RobotCode          string                  `json:"robotCode"`
	ChunkIndex         int                     `json:"chunkIndex"`
	Status             string                  `json:"status"`
	StartedAt          time.Time               `json:"startedAt"`
	EndedAt            time.Time               `json:"endedAt"`
	DurationSeconds    int                     `json:"durationSeconds"`
	ManifestObjectKey  string                  `json:"manifestObjectKey"`
	MediaObjectKeys    map[string]string       `json:"mediaObjectKeys"`
	AvailableFileTypes map[string]bool         `json:"availableFileTypes,omitempty"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
	Files              []RecordingFileResponse `json:"files,omitempty"`
}

type RecordingFileResponse struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	ContentType string `json:"contentType"`
	ObjectKey   string `json:"objectKey,omitempty"`
	URL         string `json:"url,omitempty"`
}

type RecordingTickResponse struct {
	Chunk    RecordingChunkResponse `json:"chunk"`
	Manifest map[string]any         `json:"manifest"`
}

func Robot(robot domain.Robot) RobotResponse {
	return RobotResponse(robot)
}

func Robots(robots []domain.Robot) []RobotResponse {
	response := make([]RobotResponse, 0, len(robots))
	for _, robot := range robots {
		response = append(response, Robot(robot))
	}
	return response
}

func RobotConnectionInfo(info domain.RobotConnectionInfo) RobotConnectionInfoResponse {
	return RobotConnectionInfoResponse(info)
}

func Mission(mission domain.Mission) MissionResponse {
	return MissionResponse{
		ID:          mission.ID,
		MissionCode: mission.MissionCode,
		Name:        mission.Name,
		MissionType: mission.MissionType,
		Status:      mission.Status,
		SiteNote:    mission.SiteNote,
		RobotCode:   mission.RobotCode,
		RobotCodes:  append([]string(nil), mission.RobotCodes...),
		StartedAt:   mission.StartedAt,
		EndedAt:     mission.EndedAt,
		CreatedAt:   mission.CreatedAt,
		UpdatedAt:   mission.UpdatedAt,
	}
}

func Missions(missions []domain.Mission) []MissionResponse {
	response := make([]MissionResponse, 0, len(missions))
	for _, mission := range missions {
		response = append(response, Mission(mission))
	}
	return response
}

func StreamingStatus(status domain.StreamingStatus) StreamingStatusResponse {
	tracks := make([]StreamingTrackResponse, 0, len(status.PublishedTracks))
	for _, track := range status.PublishedTracks {
		tracks = append(tracks, StreamingTrackResponse(track))
	}
	return StreamingStatusResponse{
		RobotCode:             status.RobotCode,
		MissionID:             status.MissionID,
		RoomID:                status.RoomID,
		Status:                status.Status,
		PublishedTracks:       tracks,
		PublishedDataChannels: append([]string(nil), status.PublishedDataChannels...),
		SentAt:                status.SentAt,
	}
}

func StreamingStatuses(statuses []domain.StreamingStatus) []StreamingStatusResponse {
	response := make([]StreamingStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		response = append(response, StreamingStatus(status))
	}
	return response
}

func Telemetry(snapshot domain.TelemetrySnapshot) TelemetryResponse {
	return TelemetryResponse(snapshot)
}

func TelemetryList(snapshots []domain.TelemetrySnapshot) []TelemetryResponse {
	response := make([]TelemetryResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		response = append(response, Telemetry(snapshot))
	}
	return response
}

func SensorReading(reading domain.SensorReading) SensorReadingResponse {
	return SensorReadingResponse(reading)
}

func SensorReadings(readings []domain.SensorReading) []SensorReadingResponse {
	response := make([]SensorReadingResponse, 0, len(readings))
	for _, reading := range readings {
		response = append(response, SensorReading(reading))
	}
	return response
}

func RecordingChunk(chunk domain.RecordingChunk) RecordingChunkResponse {
	return RecordingChunkResponse{
		ID:                 chunk.ID,
		RecordingSessionID: chunk.RecordingSessionID,
		MissionID:          chunk.MissionID,
		MissionCode:        chunk.MissionCode,
		RobotCode:          chunk.RobotCode,
		ChunkIndex:         chunk.ChunkIndex,
		Status:             chunk.Status,
		StartedAt:          chunk.StartedAt,
		EndedAt:            chunk.EndedAt,
		DurationSeconds:    chunk.DurationSeconds,
		ManifestObjectKey:  chunk.ManifestObjectKey,
		MediaObjectKeys:    copyStringMap(chunk.MediaObjectKeys),
		AvailableFileTypes: copyBoolMap(chunk.AvailableFileTypes),
		CreatedAt:          chunk.CreatedAt,
		UpdatedAt:          chunk.UpdatedAt,
	}
}

func RecordingTick(result domain.RecordingTickResult) RecordingTickResponse {
	return RecordingTickResponse{
		Chunk:    RecordingChunk(result.Chunk),
		Manifest: result.Manifest,
	}
}

func copyStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func copyBoolMap(input map[string]bool) map[string]bool {
	if input == nil {
		return nil
	}
	output := make(map[string]bool, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
