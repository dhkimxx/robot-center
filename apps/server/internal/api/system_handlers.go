package api

import (
	"context"
	"errors"
	"net/http"
	"robot-center/apps/server/internal/service"
	"strings"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "app-server",
		"startedAt": s.started.Format(time.RFC3339),
	})
}

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
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "app-server",
		"status":  "ok",
		"components": []map[string]string{
			{"name": "app-server", "status": "ok"},
			{"name": "recorder-worker", "status": s.componentHTTPStatus(requestContext, s.config.RecorderWorkerURL+"/healthz")},
			{"name": "turn", "status": "configured"},
			{"name": "postgres", "status": "configured"},
			{"name": "minio", "status": "configured"},
		},
		"config": map[string]string{
			"environment":   s.config.Environment,
			"publicUrl":     s.config.PublicURL,
			"minioEndpoint": s.config.MinIOEndpoint,
			"minioBucket":   s.config.MinIOBucket,
		},
		"objectStorage": s.readObjectStorageStatus(requestContext),
		"summary": map[string]int{
			"robots":     len(robots),
			"missions":   len(missions),
			"sfuRooms":   len(sfuRooms),
			"recordings": len(recordings),
		},
		"sfuRooms": sfuRooms,
	})
}

func (s *Server) readObjectStorageStatus(ctx context.Context) any {
	if s.services.Storage == nil {
		return map[string]any{
			"status": "unavailable",
			"bucket": s.config.MinIOBucket,
		}
	}
	usage, err := s.services.Storage.GetObjectStorageUsage(ctx)
	if err != nil {
		return map[string]any{
			"status": "unavailable",
			"bucket": s.config.MinIOBucket,
			"error":  err.Error(),
		}
	}
	return usage
}

type clearObjectStorageRequest struct {
	Confirmation string `json:"confirmation"`
}

func (s *Server) handleClearObjectStorage(w http.ResponseWriter, r *http.Request) {
	var request clearObjectStorageRequest
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

	writeJSON(w, http.StatusOK, map[string]any{
		"objectStorage": result,
	})
}

func (s *Server) handleRTCConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":                 "sfu",
		"signalingUrl":         s.config.SFUOperatorWebSocketURL(),
		"operatorSignalingUrl": s.config.SFUOperatorWebSocketURL(),
		"iceTransportPolicy":   "relay",
		"iceServers": []map[string]any{
			{
				"urls":       []string{s.config.TURNPublicURL},
				"username":   s.config.TURNUsername,
				"credential": s.config.TURNPassword,
			},
		},
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
