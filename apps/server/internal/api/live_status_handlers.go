package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/service"
	"robot-center/apps/server/internal/store"
	"strings"
	"time"
)

// @Summary 임무 live status 조회
// @Description 임무에 배정된 로봇의 연결, 스트림, 녹화 상태를 반환합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "임무 코드"
// @Success 200 {object} dto.MissionLiveStatusResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/live-status [get]
func (s *Server) handleMissionLiveStatus(w http.ResponseWriter, r *http.Request) {
	missionCode := strings.TrimSpace(r.PathValue("missionCode"))
	missions, err := s.services.Missions.ListMissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	mission, found := findMissionByCode(missions, missionCode)
	if !found {
		writeError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}
	robots, err := s.services.Robots.ListRobots(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recordings, err := s.services.Recording.ListRecordingChunks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	streamSessions, err := s.services.Streams.ListRobotStreamSessionsForMission(r.Context(), mission.MissionCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	liveStatus := s.services.Live.BuildMissionLiveStatus(service.LiveStatusInput{
		Mission:         mission,
		Robots:          robots,
		RecordingChunks: recordings,
		StreamSessions:  streamSessions,
		ObservedRooms:   s.sfuHub.ObservedRooms(),
		Recorder:        s.fetchRecorderRuntimeSnapshot(r.Context()),
		Now:             time.Now().UTC(),
		FreshnessWindow: 30 * time.Second,
	})
	writeJSON(w, http.StatusOK, dto.MissionLiveStatus(liveStatus))
}

type recorderHealthResponse struct {
	Subscriber recorderSubscriberHealth `json:"subscriber"`
}

type recorderSubscriberHealth struct {
	Rooms []recorderRoomHealth `json:"rooms"`
}

type recorderRoomHealth struct {
	RoomID      string                `json:"roomId"`
	MissionCode string                `json:"missionCode"`
	Robots      []recorderRobotHealth `json:"robots"`
}

type recorderRobotHealth struct {
	RobotCode        string    `json:"robotCode"`
	TrackCount       int       `json:"trackCount"`
	DataChannelCount int       `json:"dataChannelCount"`
	LastTrackAt      time.Time `json:"lastTrackAt"`
	LastDataAt       time.Time `json:"lastDataAt"`
	LastPersistedAt  time.Time `json:"lastPersistedAt"`
}

func (s *Server) fetchRecorderRuntimeSnapshot(ctx context.Context) service.RecorderRuntimeSnapshot {
	targetURL := strings.TrimSpace(s.config.RecorderWorkerInternalURL)
	if targetURL == "" {
		return service.RecorderRuntimeSnapshot{}
	}
	recorderURL, err := url.JoinPath(targetURL, "healthz")
	if err != nil {
		return service.RecorderRuntimeSnapshot{}
	}
	requestContext, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, recorderURL, nil)
	if err != nil {
		return service.RecorderRuntimeSnapshot{}
	}
	response, err := (&http.Client{Timeout: 500 * time.Millisecond}).Do(request)
	if err != nil {
		return service.RecorderRuntimeSnapshot{}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return service.RecorderRuntimeSnapshot{}
	}
	var payload recorderHealthResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return service.RecorderRuntimeSnapshot{}
	}
	return service.RecorderRuntimeSnapshot{
		Available: true,
		Rooms:     recorderRuntimeRooms(payload.Subscriber.Rooms),
	}
}

func recorderRuntimeRooms(rooms []recorderRoomHealth) []service.RecorderRoomRuntime {
	output := make([]service.RecorderRoomRuntime, 0, len(rooms))
	for _, room := range rooms {
		robots := make([]service.RecorderRobotRuntime, 0, len(room.Robots))
		for _, robot := range room.Robots {
			robots = append(robots, service.RecorderRobotRuntime{
				RobotCode:        strings.TrimSpace(robot.RobotCode),
				TrackCount:       robot.TrackCount,
				DataChannelCount: robot.DataChannelCount,
				LastTrackAt:      nonZeroTimePointer(robot.LastTrackAt),
				LastDataAt:       nonZeroTimePointer(robot.LastDataAt),
				LastPersistedAt:  nonZeroTimePointer(robot.LastPersistedAt),
			})
		}
		output = append(output, service.RecorderRoomRuntime{
			RoomID:      strings.TrimSpace(room.RoomID),
			MissionCode: strings.TrimSpace(room.MissionCode),
			Robots:      robots,
		})
	}
	return output
}

func nonZeroTimePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func findMissionByCode(missions []domain.Mission, missionCode string) (domain.Mission, bool) {
	for _, mission := range missions {
		if mission.MissionCode == missionCode {
			return mission, true
		}
	}
	return domain.Mission{}, false
}
