package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
)

type MissionLiveStatusResponse struct {
	MissionCode   string                    `json:"missionCode"`
	MissionStatus string                    `json:"missionStatus"`
	ObservedAt    time.Time                 `json:"observedAt"`
	Robots        []RobotLiveStatusResponse `json:"robots"`
}

type RobotLiveStatusResponse struct {
	RobotCode   string                       `json:"robotCode"`
	DisplayName string                       `json:"displayName"`
	Connection  LiveConnectionStatusResponse `json:"connection"`
	Stream      LiveStreamStatusResponse     `json:"stream"`
	Recording   LiveRecordingStatusResponse  `json:"recording"`
}

type LiveConnectionStatusResponse struct {
	State      string     `json:"state"`
	Source     string     `json:"source"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
}

type LiveStreamStatusResponse struct {
	State            string                         `json:"state"`
	Source           string                         `json:"source"`
	RoomID           string                         `json:"roomId"`
	TrackCount       int                            `json:"trackCount"`
	DataChannelCount int                            `json:"dataChannelCount"`
	StartedAt        *time.Time                     `json:"startedAt,omitempty"`
	LastTrackAt      *time.Time                     `json:"lastTrackAt,omitempty"`
	LastDataAt       *time.Time                     `json:"lastDataAt,omitempty"`
	LastMediaAt      *time.Time                     `json:"lastMediaAt,omitempty"`
	Diagnostics      *LiveStreamDiagnosticsResponse `json:"diagnostics,omitempty"`
	Reason           string                         `json:"reason,omitempty"`
}

type LiveStreamDiagnosticsResponse struct {
	LastSessionMediaAt *time.Time `json:"lastSessionMediaAt,omitempty"`
	PreviousEndedAt    *time.Time `json:"previousEndedAt,omitempty"`
	ReconnectCount     int        `json:"reconnectCount"`
}

type LiveRecordingStatusResponse struct {
	State             string                             `json:"state"`
	Source            string                             `json:"source"`
	LatestChunk       *LiveRecordingChunkSummaryResponse `json:"latestChunk,omitempty"`
	LatestChunkID     string                             `json:"latestChunkId,omitempty"`
	LatestChunkStatus string                             `json:"latestChunkStatus,omitempty"`
	Reason            string                             `json:"reason,omitempty"`
}

type LiveRecordingChunkSummaryResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ChunkIndex int       `json:"chunkIndex"`
}

func MissionLiveStatus(status domain.MissionLiveStatus) MissionLiveStatusResponse {
	robots := make([]RobotLiveStatusResponse, 0, len(status.Robots))
	for _, robot := range status.Robots {
		robots = append(robots, RobotLiveStatusResponse{
			RobotCode:   robot.RobotCode,
			DisplayName: robot.DisplayName,
			Connection:  LiveConnectionStatusResponse(robot.Connection),
			Stream:      liveStreamStatus(robot.Stream),
			Recording:   liveRecordingStatus(robot.Recording),
		})
	}
	return MissionLiveStatusResponse{
		MissionCode:   status.MissionCode,
		MissionStatus: status.MissionStatus,
		ObservedAt:    status.ObservedAt,
		Robots:        robots,
	}
}

func liveStreamStatus(status domain.LiveStreamStatus) LiveStreamStatusResponse {
	response := LiveStreamStatusResponse{
		State:            status.State,
		Source:           status.Source,
		RoomID:           status.RoomID,
		TrackCount:       status.TrackCount,
		DataChannelCount: status.DataChannelCount,
		StartedAt:        status.StartedAt,
		LastTrackAt:      status.LastTrackAt,
		LastDataAt:       status.LastDataAt,
		LastMediaAt:      status.LastMediaAt,
		Reason:           status.Reason,
	}
	if status.Diagnostics != nil {
		response.Diagnostics = &LiveStreamDiagnosticsResponse{
			LastSessionMediaAt: status.Diagnostics.LastSessionMediaAt,
			PreviousEndedAt:    status.Diagnostics.PreviousEndedAt,
			ReconnectCount:     status.Diagnostics.ReconnectCount,
		}
	}
	return response
}

func liveRecordingStatus(status domain.LiveRecordingStatus) LiveRecordingStatusResponse {
	response := LiveRecordingStatusResponse{
		State:             status.State,
		Source:            status.Source,
		LatestChunkID:     status.LatestChunkID,
		LatestChunkStatus: status.LatestChunkStatus,
		Reason:            status.Reason,
	}
	if status.LatestChunk != nil {
		response.LatestChunk = &LiveRecordingChunkSummaryResponse{
			ID:         status.LatestChunk.ID,
			Status:     status.LatestChunk.Status,
			StartedAt:  status.LatestChunk.StartedAt,
			EndedAt:    status.LatestChunk.EndedAt,
			UpdatedAt:  status.LatestChunk.UpdatedAt,
			ChunkIndex: status.LatestChunk.ChunkIndex,
		}
	}
	return response
}
