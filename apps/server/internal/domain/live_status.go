package domain

import "time"

type MissionLiveStatus struct {
	MissionCode   string            `json:"missionCode"`
	MissionStatus string            `json:"missionStatus"`
	ObservedAt    time.Time         `json:"observedAt"`
	Robots        []RobotLiveStatus `json:"robots"`
}

type RobotLiveStatus struct {
	RobotCode   string               `json:"robotCode"`
	DisplayName string               `json:"displayName"`
	Connection  LiveConnectionStatus `json:"connection"`
	Stream      LiveStreamStatus     `json:"stream"`
	Recording   LiveRecordingStatus  `json:"recording"`
}

type LiveConnectionStatus struct {
	State      string     `json:"state"`
	Source     string     `json:"source"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
}

type LiveStreamStatus struct {
	State            string                 `json:"state"`
	Source           string                 `json:"source"`
	RoomID           string                 `json:"roomId"`
	TrackCount       int                    `json:"trackCount"`
	DataChannelCount int                    `json:"dataChannelCount"`
	StartedAt        *time.Time             `json:"startedAt,omitempty"`
	LastTrackAt      *time.Time             `json:"lastTrackAt,omitempty"`
	LastDataAt       *time.Time             `json:"lastDataAt,omitempty"`
	LastMediaAt      *time.Time             `json:"lastMediaAt,omitempty"`
	Diagnostics      *LiveStreamDiagnostics `json:"diagnostics,omitempty"`
	Reason           string                 `json:"reason,omitempty"`
}

type LiveStreamDiagnostics struct {
	LastSessionMediaAt *time.Time `json:"lastSessionMediaAt,omitempty"`
	PreviousEndedAt    *time.Time `json:"previousEndedAt,omitempty"`
	ReconnectCount     int        `json:"reconnectCount"`
}

type LiveRecordingStatus struct {
	State             string                     `json:"state"`
	Source            string                     `json:"source"`
	LatestChunk       *LiveRecordingChunkSummary `json:"latestChunk,omitempty"`
	LatestChunkID     string                     `json:"latestChunkId,omitempty"`
	LatestChunkStatus string                     `json:"latestChunkStatus,omitempty"`
	Reason            string                     `json:"reason,omitempty"`
}

type LiveRecordingChunkSummary struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ChunkIndex int       `json:"chunkIndex"`
}
