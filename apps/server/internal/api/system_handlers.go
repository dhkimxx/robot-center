package api

import (
	"context"
	"errors"
	"net/http"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/service"

	"strings"
	"time"
)

func (s *Server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	requestContext := r.Context()
	robots, err := s.services.Robots.ListRobots(requestContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	missions, err := s.services.Missions.ListMissions(requestContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recordings, err := s.services.Recording.ListRecordingChunks(requestContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sfuRooms := s.sfuHub.Summaries()
	writeJSON(w, http.StatusOK, dto.SystemStatus(dto.SystemStatusInput{
		Environment:               s.config.Environment,
		AppServerPublicURL:        s.config.AppServerPublicURL,
		RecorderWorkerInternalURL: s.config.RecorderWorkerInternalURL,
		MinIOInternalURL:          s.config.MinIOInternalURL,
		MinIOPublicURL:            s.config.MinIOPublicURL,
		MinIOBucket:               s.config.MinIOBucket,
		RecorderWorkerStatus:      s.componentHTTPStatus(requestContext, s.config.RecorderWorkerInternalURL+"/healthz"),
		ObjectStorage:             s.readObjectStorageStatus(requestContext),
		RobotCount:                len(robots),
		MissionCount:              len(missions),
		RecordingCount:            len(recordings),
		SFURooms:                  sfuRooms,
	}))
}

func (s *Server) readObjectStorageStatus(ctx context.Context) dto.ObjectStorageStatusResponse {
	if s.services.Storage == nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, nil)
	}
	usage, err := s.services.Storage.GetObjectStorageUsage(ctx)
	if err != nil {
		return dto.ObjectStorageUnavailable(s.config.MinIOBucket, err)
	}
	return dto.ObjectStorageStatus(usage)
}

func (s *Server) handleClearObjectStorage(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearObjectStorageRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Storage.ClearObjectStorage(r.Context(), request.Confirmation)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSystemActionForbidden):
			writeError(w, http.StatusForbidden, err)
		case errors.Is(err, service.ErrSystemActionConfirmationRequired):
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearObjectStorageResponse{
		ObjectStorage: result,
	})
}

func (s *Server) handleClearSensorData(w http.ResponseWriter, r *http.Request) {
	var request dto.ClearSensorDataRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Sensors.ClearSensorData(r.Context(), s.config.Environment, request.Confirmation)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSystemActionForbidden):
			writeError(w, http.StatusForbidden, err)
		case errors.Is(err, service.ErrSystemActionConfirmationRequired):
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.ClearSensorDataResponse{
		SensorData: result,
	})
}

func (s *Server) componentHTTPStatus(ctx context.Context, targetURL string) string {
	if strings.TrimSpace(targetURL) == "" {
		return "unknown"
	}
	client := http.Client{Timeout: 500 * time.Millisecond}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "unreachable"
	}
	response, err := client.Do(request)
	if err != nil {
		return "unreachable"
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return "ok"
	}
	return "degraded"
}
