package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/service"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
)

type Server struct {
	config   config.AppServerConfig
	services *service.Services
	sfuHub   *sfu.Hub
	started  time.Time
}

func NewServer(cfg config.AppServerConfig) *Server {
	return NewServerWithStore(cfg, store.NewMemoryStore(cfg.PublicURL))
}

func NewServerWithStore(cfg config.AppServerConfig, repository store.Store) *Server {
	services := service.NewServices(repository)
	return &Server{
		config:   cfg,
		services: services,
		sfuHub: sfu.NewHub(sfu.Config{
			TURNURL:      cfg.TURNURL,
			TURNUsername: cfg.TURNUsername,
			TURNPassword: cfg.TURNPassword,
			ValidateRobotPublisher: func(roomID string, robotCode string) error {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				return services.Missions.ValidateActiveMissionRobot(ctx, roomID, robotCode)
			},
		}),
		started: time.Now().UTC(),
	}
}

func NewServerFromConfig(ctx context.Context, cfg config.AppServerConfig) (*Server, error) {
	switch strings.TrimSpace(cfg.StoreDriver) {
	case "", "memory":
		return NewServerWithStore(cfg, store.NewMemoryStore(cfg.PublicURL)), nil
	case "postgres":
		repository, err := store.NewPostgresStore(ctx, store.PostgresConfig{
			DSN:         cfg.PostgresDSN,
			ServerURL:   cfg.PublicURL,
			MinIOBucket: cfg.MinIOBucket,
		})
		if err != nil {
			return nil, err
		}
		return NewServerWithStore(cfg, repository), nil
	default:
		return nil, fmt.Errorf("unsupported STORE_DRIVER %q", cfg.StoreDriver)
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/system/status", s.handleSystemStatus)
	mux.HandleFunc("GET /api/rtc-config", s.handleRTCConfig)
	mux.HandleFunc("GET /api/recording-targets", s.handleRecordingTargets)
	mux.HandleFunc("GET /api/streaming-statuses", s.handleListStreamingStatuses)
	mux.HandleFunc("GET /api/sensor-descriptors", s.handleListSensorDescriptors)
	mux.HandleFunc("POST /api/sensor-descriptors", s.handleCreateSensorSamples)
	mux.HandleFunc("GET /api/sensor-samples", s.handleListSensorSamples)
	mux.HandleFunc("POST /api/sensor-samples", s.handleCreateSensorSamples)
	mux.HandleFunc("GET /api/sensor-latest", s.handleListSensorLatest)
	mux.HandleFunc("GET /api/recordings", s.handleListRecordings)
	mux.HandleFunc("POST /api/recorder/tick", s.handleRecorderTick)
	mux.HandleFunc("POST /api/recorder/chunks/{chunkID}/uploaded", s.handleRecorderChunkUploaded)
	mux.HandleFunc("POST /api/recorder/chunks/{chunkID}/files/{fileType}/uploaded", s.handleRecorderFileUploaded)
	mux.HandleFunc("GET /api/robots", s.handleListRobots)
	mux.HandleFunc("POST /api/robots", s.handleCreateRobot)
	mux.HandleFunc("PATCH /api/robots/{robotCode}", s.handleUpdateRobot)
	mux.HandleFunc("DELETE /api/robots/{robotCode}", s.handleArchiveRobot)
	mux.HandleFunc("GET /api/robots/{robotCode}/connection-info", s.handleGetRobotConnectionInfo)
	mux.HandleFunc("POST /api/robots/{robotCode}/connection-token", s.handleRotateRobotConnectionToken)
	mux.HandleFunc("GET /api/missions", s.handleListMissions)
	mux.HandleFunc("POST /api/missions", s.handleCreateMission)
	mux.HandleFunc("POST /api/missions/{missionCode}/start", s.handleStartMission)
	mux.HandleFunc("POST /api/missions/{missionCode}/end", s.handleEndMission)
	mux.HandleFunc("POST /api/robot-gateway/heartbeat", s.handleRobotGatewayHeartbeat)
	mux.HandleFunc("GET /api/robot-gateway/mission", s.handleRobotGatewayMission)
	mux.HandleFunc("POST /api/robot-gateway/streaming-status", s.handleRobotGatewayStreamingStatus)
	mux.HandleFunc("GET /sfu/ws", s.handleSFUWebSocket)

	if s.config.WebStaticDir != "" {
		mux.Handle("/", s.staticHandler())
	}

	return withRequestHeaders(mux)
}

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
		"summary": map[string]int{
			"robots":     len(robots),
			"missions":   len(missions),
			"sfuRooms":   len(sfuRooms),
			"recordings": len(recordings),
		},
		"sfuRooms": sfuRooms,
	})
}

func (s *Server) handleRTCConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":               "sfu",
		"signalingUrl":       s.config.SFUWebSocketURL,
		"iceTransportPolicy": "relay",
		"iceServers": []map[string]any{
			{
				"urls":       []string{s.config.TURNURL},
				"username":   s.config.TURNUsername,
				"credential": s.config.TURNPassword,
			},
		},
	})
}

func (s *Server) handleRecordingTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.services.Missions.RecordingTargets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"targets": dto.Missions(targets),
	})
}

func (s *Server) handleListStreamingStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := s.services.Streaming.ListStreamingStatuses(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"streamingStatuses": dto.StreamingStatuses(statuses),
	})
}

func (s *Server) handleListSensorDescriptors(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	descriptors, err := s.services.Sensors.ListSensorDescriptors(r.Context(), missionID, robotCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sensorDescriptors": dto.SensorDescriptors(descriptors),
	})
}

func (s *Server) handleListSensorSamples(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	sensorID := strings.TrimSpace(r.URL.Query().Get("sensorId"))
	limit := intQueryValue(r, "limit", 100)
	samples, err := s.services.Sensors.ListSensorSamples(r.Context(), missionID, robotCode, sensorID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sensorSamples": dto.SensorSamples(samples),
	})
}

func (s *Server) handleListSensorLatest(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	latest, err := s.services.Sensors.ListLatestSensorSamples(r.Context(), missionID, robotCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"missionId": missionID,
		"robotCode": robotCode,
		"sensors":   dto.SensorLatest(latest),
	})
}

func (s *Server) handleCreateSensorSamples(w http.ResponseWriter, r *http.Request) {
	envelope, err := decodeSensorEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	samples, err := s.services.Sensors.SaveSensorEnvelope(r.Context(), envelope)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"sensorSamples": dto.SensorSamples(samples),
	})
}

func (s *Server) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	recordings, err := s.services.Recording.ListRecordingChunks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	response := make([]dto.RecordingChunkResponse, 0, len(recordings))
	for _, recording := range recordings {
		response = append(response, s.createRecordingResponse(recording))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"recordings": response,
	})
}

func (s *Server) createRecordingResponse(recording domain.RecordingChunk) dto.RecordingChunkResponse {
	response := dto.RecordingChunk(recording)
	response.Files = []dto.RecordingFileResponse{
		s.createRecordingFileResponse(recording, "rgb_audio_mp4", "RGB MP4", "video/mp4", recording.MediaObjectKeys["rgbMp4"], recording.AvailableFileTypes["rgb_audio_mp4"]),
		s.createRecordingFileResponse(recording, "thermal_mp4", "Thermal MP4", "video/mp4", recording.MediaObjectKeys["thermal"], recording.AvailableFileTypes["thermal_mp4"]),
		s.createRecordingFileResponse(recording, "sensor_jsonl", "Sensor JSONL", "application/x-ndjson", recording.MediaObjectKeys["sensor"], recording.AvailableFileTypes["sensor_jsonl"]),
		s.createRecordingFileResponse(recording, "telemetry_jsonl", "Telemetry/GPS JSONL", "application/x-ndjson", recording.MediaObjectKeys["telemetry"], recording.AvailableFileTypes["telemetry_jsonl"]),
		s.createRecordingFileResponse(recording, "manifest", "저장 메타데이터", "application/json", recording.ManifestObjectKey, recording.AvailableFileTypes["manifest"] || recording.Status == "uploaded"),
	}
	return response
}

func (s *Server) createRecordingFileResponse(recording domain.RecordingChunk, fileType string, label string, contentType string, objectKey string, available bool) dto.RecordingFileResponse {
	status := "planned"
	fileURL := ""
	if available {
		status = "available"
		fileURL = s.createStorageObjectURL(objectKey)
	} else if recording.Status == "recording" {
		status = "recording"
	}
	return dto.RecordingFileResponse{
		Type:        fileType,
		Label:       label,
		Status:      status,
		ContentType: contentType,
		ObjectKey:   objectKey,
		URL:         fileURL,
	}
}

func (s *Server) createStorageObjectURL(objectKey string) string {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return ""
	}

	publicURL, publicErr := url.Parse(s.config.PublicURL)
	minioURL, minioErr := url.Parse(s.config.MinIOEndpoint)

	scheme := "http"
	if publicErr == nil && publicURL.Scheme != "" {
		scheme = publicURL.Scheme
	} else if minioErr == nil && minioURL.Scheme != "" {
		scheme = minioURL.Scheme
	}

	host := ""
	if publicErr == nil {
		host = publicURL.Hostname()
	}
	if host == "" && minioErr == nil {
		host = minioURL.Hostname()
	}
	if host == "" {
		host = "localhost"
	}

	port := "9000"
	if minioErr == nil && minioURL.Port() != "" {
		port = minioURL.Port()
	}
	bucket := strings.TrimSpace(s.config.MinIOBucket)
	if bucket == "" {
		bucket = "robot-center"
	}
	encodedPath := encodeObjectPath(bucket + "/" + objectKey)
	return fmt.Sprintf("%s://%s/%s", scheme, net.JoinHostPort(host, port), encodedPath)
}

func encodeObjectPath(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

type createRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

type updateRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

func (s *Server) handleCreateRobot(w http.ResponseWriter, r *http.Request) {
	var request createRobotRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, connectionInfo, err := s.services.Robots.CreateRobot(r.Context(), store.CreateRobotInput{
		DisplayName: strings.TrimSpace(request.DisplayName),
		ModelName:   strings.TrimSpace(request.ModelName),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"robot":          dto.Robot(robot),
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}

func (s *Server) handleUpdateRobot(w http.ResponseWriter, r *http.Request) {
	var request updateRobotRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, err := s.services.Robots.UpdateRobot(r.Context(), r.PathValue("robotCode"), store.UpdateRobotInput{
		DisplayName: strings.TrimSpace(request.DisplayName),
		ModelName:   strings.TrimSpace(request.ModelName),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"robot": dto.Robot(robot),
	})
}

func (s *Server) handleArchiveRobot(w http.ResponseWriter, r *http.Request) {
	robot, err := s.services.Robots.ArchiveRobot(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"robot": dto.Robot(robot),
	})
}

func (s *Server) handleListRobots(w http.ResponseWriter, r *http.Request) {
	robots, err := s.services.Robots.ListRobots(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"robots": dto.Robots(robots),
	})
}

func (s *Server) handleGetRobotConnectionInfo(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.GetRobotConnectionInfo(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}

func (s *Server) handleRotateRobotConnectionToken(w http.ResponseWriter, r *http.Request) {
	connectionInfo, err := s.services.Robots.RotateRobotConnectionToken(r.Context(), r.PathValue("robotCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connectionInfo": dto.RobotConnectionInfo(connectionInfo),
	})
}

type createMissionRequest struct {
	Name        string   `json:"name"`
	MissionType string   `json:"missionType"`
	SiteNote    string   `json:"siteNote"`
	RobotCode   string   `json:"robotCode"`
	RobotCodes  []string `json:"robotCodes"`
}

func (s *Server) handleCreateMission(w http.ResponseWriter, r *http.Request) {
	var request createMissionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !validMissionType(request.MissionType) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid missionType %q", request.MissionType))
		return
	}

	mission, err := s.services.Missions.CreateMission(r.Context(), store.CreateMissionInput{
		Name:        strings.TrimSpace(request.Name),
		MissionType: strings.TrimSpace(request.MissionType),
		SiteNote:    strings.TrimSpace(request.SiteNote),
		RobotCode:   strings.TrimSpace(request.RobotCode),
		RobotCodes:  request.RobotCodes,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"mission": dto.Mission(mission),
	})
}

func (s *Server) handleListMissions(w http.ResponseWriter, r *http.Request) {
	missions, err := s.services.Missions.ListMissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"missions": dto.Missions(missions),
	})
}

func (s *Server) handleStartMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.StartMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mission": dto.Mission(mission),
	})
}

func (s *Server) handleEndMission(w http.ResponseWriter, r *http.Request) {
	mission, err := s.services.Missions.EndMission(r.Context(), r.PathValue("missionCode"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.sfuHub.CloseRoom(mission.MissionCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"mission": dto.Mission(mission),
	})
}

type heartbeatRequest struct {
	RobotCode      string    `json:"robotCode"`
	State          string    `json:"state"`
	BatteryPercent int       `json:"batteryPercent"`
	NetworkQuality string    `json:"networkQuality"`
	SentAt         time.Time `json:"sentAt"`
}

func (s *Server) handleRobotGatewayHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request heartbeatRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	robot, err := s.services.Robots.ApplyHeartbeat(r.Context(), store.HeartbeatInput{
		RobotCode:      strings.TrimSpace(request.RobotCode),
		State:          valueOrDefault(strings.TrimSpace(request.State), "online"),
		BatteryPercent: request.BatteryPercent,
		NetworkQuality: request.NetworkQuality,
		SentAt:         request.SentAt,
	}, bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"robotId":    robot.ID,
		"robotCode":  robot.RobotCode,
		"status":     robot.Status,
		"serverTime": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) handleRobotGatewayMission(w http.ResponseWriter, r *http.Request) {
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	mission, found, err := s.services.Missions.FindActiveMissionForRobot(r.Context(), robotCode, bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, map[string]any{
			"missionId":     nil,
			"missionStatus": "none",
		})
		return
	}

	roomID := mission.MissionCode
	legacyRoomID := domain.StreamRoomID(mission.MissionCode, robotCode)
	writeJSON(w, http.StatusOK, map[string]any{
		"missionId":     mission.ID,
		"missionCode":   mission.MissionCode,
		"missionStatus": mission.Status,
		"roomId":        roomID,
		"legacyRoomId":  legacyRoomID,
		"sfu": map[string]any{
			"signalingUrl":       s.config.SFUWebSocketURL + "?room=" + url.QueryEscape(roomID) + "&role=robot&robotCode=" + url.QueryEscape(robotCode),
			"iceTransportPolicy": "relay",
		},
		"turnServers": []map[string]any{
			{
				"urls":       []string{s.config.TURNURL},
				"username":   s.config.TURNUsername,
				"credential": s.config.TURNPassword,
			},
		},
		"tracks": []string{
			sfu.StreamRoleTrackVideo1,
			sfu.StreamRoleTrackVideo2,
			sfu.StreamRoleTrackAudio1,
			sfu.StreamRoleTrackAudio2,
		},
		"dataChannels": []string{
			sfu.StreamRoleChannelTelemetry,
			sfu.StreamRoleChannelSpatial,
			sfu.StreamRoleChannelEvent,
			sfu.StreamRoleChannelControl,
		},
		"legacyTracks":       []string{"rgb", "thermal", "audio"},
		"legacyDataChannels": []string{"sensor", "telemetry"},
		"videoPolicy": map[string]string{
			"mode": "robot_defined",
		},
	})
}

func (s *Server) handleRobotGatewayStreamingStatus(w http.ResponseWriter, r *http.Request) {
	var request domain.StreamingStatus
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.SentAt.IsZero() {
		request.SentAt = time.Now().UTC()
	}

	robot, err := s.services.Streaming.ApplyStreamingStatus(r.Context(), request, bearerToken(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accepted":   true,
		"robotCode":  robot.RobotCode,
		"status":     robot.Status,
		"serverTime": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

type recorderTickRequest struct {
	MissionCode          string    `json:"missionCode"`
	RobotCode            string    `json:"robotCode"`
	ChunkDurationSeconds int       `json:"chunkDurationSeconds"`
	TickAt               time.Time `json:"tickAt"`
}

type recorderUploadRequest struct {
	SizeBytes *int64 `json:"sizeBytes"`
	Checksum  string `json:"checksum"`
}

func (s *Server) handleRecorderTick(w http.ResponseWriter, r *http.Request) {
	var request recorderTickRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.services.Recording.ApplyRecordingTick(r.Context(), store.RecordingTickInput{
		MissionCode:          strings.TrimSpace(request.MissionCode),
		RobotCode:            strings.TrimSpace(request.RobotCode),
		ChunkDurationSeconds: request.ChunkDurationSeconds,
		TickAt:               request.TickAt,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.RecordingTick(result))
}

func (s *Server) handleRecorderChunkUploaded(w http.ResponseWriter, r *http.Request) {
	uploadMetadata, err := decodeRecorderUploadMetadata(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	chunk, err := s.services.Recording.MarkRecordingChunkUploaded(r.Context(), r.PathValue("chunkID"), uploadMetadata)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chunk": s.createRecordingResponse(chunk),
	})
}

func (s *Server) handleRecorderFileUploaded(w http.ResponseWriter, r *http.Request) {
	uploadMetadata, err := decodeRecorderUploadMetadata(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	chunk, err := s.services.Recording.MarkRecordingFileUploaded(r.Context(), r.PathValue("chunkID"), r.PathValue("fileType"), uploadMetadata)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chunk": s.createRecordingResponse(chunk),
	})
}

func (s *Server) handleSFUWebSocket(w http.ResponseWriter, r *http.Request) {
	s.sfuHub.ServeHTTP(w, r)
}

func decodeRecorderUploadMetadata(r *http.Request) (store.RecordingUploadMetadata, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return store.RecordingUploadMetadata{}, err
	}
	if len(strings.TrimSpace(string(rawPayload))) == 0 {
		return store.RecordingUploadMetadata{}, nil
	}

	var request recorderUploadRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return store.RecordingUploadMetadata{}, err
	}
	return store.RecordingUploadMetadata{
		SizeBytes: request.SizeBytes,
		Checksum:  strings.TrimSpace(request.Checksum),
	}, nil
}

func (s *Server) staticHandler() http.Handler {
	fileServer := http.FileServer(http.Dir(s.config.WebStaticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/sfu/") {
			http.NotFound(w, r)
			return
		}

		requestPath := filepath.Clean(r.URL.Path)
		if requestPath == "." || requestPath == "/" {
			requestPath = "index.html"
		}

		fullPath := filepath.Join(s.config.WebStaticDir, requestPath)
		if _, err := os.Stat(fullPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.ServeFile(w, r, filepath.Join(s.config.WebStaticDir, "index.html"))
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}

func writeStoreError(w http.ResponseWriter, err error) {
	var missionConflict *store.MissionStartConflictError
	switch {
	case errors.As(err, &missionConflict):
		conflicts := make([]map[string]string, 0, len(missionConflict.Conflicts))
		for _, conflict := range missionConflict.Conflicts {
			conflicts = append(conflicts, map[string]string{
				"robotCode":         conflict.RobotCode,
				"activeMissionCode": conflict.ActiveMissionCode,
			})
		}
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":     err.Error(),
			"conflicts": conflicts,
		})
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, store.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, store.ErrInvalidState):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusBadRequest, err)
	}
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

type sensorDescriptorRequest struct {
	SensorID     string         `json:"sensorId"`
	ChannelRole  string         `json:"channelRole"`
	DisplayName  string         `json:"displayName"`
	Kind         string         `json:"kind"`
	SensorType   string         `json:"sensorType"`
	ValueType    string         `json:"valueType"`
	Unit         string         `json:"unit"`
	SamplingRate *float64       `json:"samplingRate"`
	SampleRateHz *float64       `json:"sampleRateHz"`
	Enabled      bool           `json:"enabled"`
	Metadata     map[string]any `json:"metadata"`
}

type sensorSampleRequest struct {
	SensorID     string         `json:"sensorId"`
	ChannelRole  string         `json:"channelRole"`
	MessageID    string         `json:"messageId"`
	Sequence     int64          `json:"sequence"`
	Timestamp    *time.Time     `json:"timestamp"`
	SentAt       *time.Time     `json:"sentAt"`
	NumericValue *float64       `json:"numericValue"`
	TextValue    string         `json:"textValue"`
	BoolValue    *bool          `json:"boolValue"`
	VectorValue  map[string]any `json:"vectorValue"`
	ObjectValue  map[string]any `json:"objectValue"`
	Values       any            `json:"values"`
	ObjectKey    string         `json:"objectKey"`
	RawPayload   map[string]any `json:"rawPayload"`
}

type sensorEnvelopeRequest struct {
	MessageID   string                    `json:"messageId"`
	MessageType string                    `json:"messageType"`
	RobotCode   string                    `json:"robotCode"`
	MissionID   string                    `json:"missionId"`
	ChannelRole string                    `json:"channelRole"`
	Sequence    int64                     `json:"sequence"`
	SentAt      *time.Time                `json:"sentAt"`
	Descriptors []sensorDescriptorRequest `json:"descriptors"`
	Samples     []sensorSampleRequest     `json:"samples"`
	Payload     map[string]any            `json:"payload"`
}

func decodeSensorEnvelope(r *http.Request) (domain.SensorEnvelope, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return domain.SensorEnvelope{}, err
	}
	var request sensorEnvelopeRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return domain.SensorEnvelope{}, err
	}
	request.RobotCode = strings.TrimSpace(request.RobotCode)
	request.MissionID = strings.TrimSpace(request.MissionID)
	request.ChannelRole = strings.TrimSpace(request.ChannelRole)
	if request.RobotCode == "" {
		return domain.SensorEnvelope{}, errors.New("robotCode is required")
	}
	if request.MissionID == "" {
		return domain.SensorEnvelope{}, errors.New("missionId is required")
	}
	if request.ChannelRole == "" {
		request.ChannelRole = "channel.telemetry"
	}

	envelope := domain.SensorEnvelope{
		MessageID:   strings.TrimSpace(request.MessageID),
		MessageType: strings.TrimSpace(request.MessageType),
		RobotCode:   request.RobotCode,
		MissionID:   request.MissionID,
		ChannelRole: request.ChannelRole,
		Sequence:    request.Sequence,
		SentAt:      request.SentAt,
		ReceivedAt:  time.Now().UTC(),
		RawPayload:  append(json.RawMessage(nil), rawPayload...),
		Descriptors: make([]domain.SensorDescriptor, 0, len(request.Descriptors)),
		Samples:     make([]domain.SensorSample, 0, len(request.Samples)),
	}
	for _, descriptor := range request.Descriptors {
		sensorID := strings.TrimSpace(descriptor.SensorID)
		if sensorID == "" {
			continue
		}
		sampleRateHz := descriptor.SampleRateHz
		if sampleRateHz == nil {
			sampleRateHz = descriptor.SamplingRate
		}
		envelope.Descriptors = append(envelope.Descriptors, domain.SensorDescriptor{
			MissionID:    request.MissionID,
			RobotCode:    request.RobotCode,
			SensorID:     sensorID,
			ChannelRole:  firstText(descriptor.ChannelRole, request.ChannelRole),
			DisplayName:  firstText(descriptor.DisplayName, sensorID),
			SensorType:   inferSensorType(sensorID, firstText(descriptor.SensorType, descriptor.Kind)),
			ValueType:    firstText(descriptor.ValueType, "object"),
			Unit:         strings.TrimSpace(descriptor.Unit),
			SampleRateHz: sampleRateHz,
			Enabled:      descriptor.Enabled,
			Metadata:     marshalJSONOrEmpty(descriptor.Metadata),
		})
	}
	for _, sample := range request.Samples {
		sensorID := strings.TrimSpace(sample.SensorID)
		if sensorID == "" {
			continue
		}
		sentAt := sample.SentAt
		if sentAt == nil {
			sentAt = sample.Timestamp
		}
		if sentAt == nil {
			sentAt = request.SentAt
		}
		envelope.Samples = append(envelope.Samples, domain.SensorSample{
			MissionID:    request.MissionID,
			RobotCode:    request.RobotCode,
			SensorID:     sensorID,
			ChannelRole:  firstText(sample.ChannelRole, request.ChannelRole),
			MessageID:    firstText(sample.MessageID, request.MessageID),
			Sequence:     firstNonZero(sample.Sequence, request.Sequence),
			SentAt:       sentAt,
			ReceivedAt:   envelope.ReceivedAt,
			NumericValue: sample.NumericValue,
			TextValue:    strings.TrimSpace(sample.TextValue),
			BoolValue:    sample.BoolValue,
			VectorValue:  marshalJSONOrNil(sample.VectorValue),
			ObjectValue:  marshalSensorSampleObjectValue(sample),
			ObjectKey:    strings.TrimSpace(sample.ObjectKey),
			RawPayload:   marshalJSONOrEmpty(sample),
		})
	}
	if len(envelope.Descriptors) == 0 && len(envelope.Samples) == 0 && len(request.Payload) > 0 {
		envelope.Descriptors = append(envelope.Descriptors, domain.SensorDescriptor{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    "legacy.payload_1",
			ChannelRole: request.ChannelRole,
			DisplayName: "Legacy Payload",
			SensorType:  "legacy",
			ValueType:   "object",
			Enabled:     true,
			Metadata:    []byte("{}"),
		})
		envelope.Samples = append(envelope.Samples, domain.SensorSample{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    "legacy.payload_1",
			ChannelRole: request.ChannelRole,
			MessageID:   request.MessageID,
			Sequence:    request.Sequence,
			SentAt:      request.SentAt,
			ReceivedAt:  envelope.ReceivedAt,
			ObjectValue: marshalJSONOrEmpty(request.Payload),
			RawPayload:  envelope.RawPayload,
		})
	}
	return envelope, nil
}

func intQueryValue(r *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func firstText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func inferSensorType(sensorID string, explicitType string) string {
	if strings.TrimSpace(explicitType) != "" {
		return strings.TrimSpace(explicitType)
	}
	sensorID = strings.ToLower(strings.TrimSpace(sensorID))
	switch {
	case strings.Contains(sensorID, "position"):
		return "position"
	case strings.Contains(sensorID, "imu"):
		return "imu"
	case strings.Contains(sensorID, "odometry"):
		return "odometry"
	case strings.Contains(sensorID, "point_cloud"):
		return "point_cloud"
	case strings.Contains(sensorID, "battery"):
		return "battery"
	case strings.Contains(sensorID, "network"):
		return "network"
	case strings.Contains(sensorID, "temperature"):
		return "temperature"
	case strings.Contains(sensorID, "humidity"):
		return "humidity"
	case strings.Contains(sensorID, "gas"):
		return "gas"
	default:
		return "unknown"
	}
}

func marshalSensorSampleObjectValue(sample sensorSampleRequest) json.RawMessage {
	if sample.ObjectValue != nil {
		return marshalJSONOrNil(sample.ObjectValue)
	}
	if sample.Values != nil {
		return marshalJSONOrNil(sample.Values)
	}
	return nil
}

func marshalJSONOrEmpty(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return []byte("{}")
	}
	return payload
}

func marshalJSONOrNil(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return nil
	}
	return payload
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimPrefix(header, "Bearer ")
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func validMissionType(missionType string) bool {
	switch missionType {
	case "mountain_rescue", "collapse_site", "underground_facility":
		return true
	default:
		return false
	}
}

func withRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robot-Center", "app-server")
		next.ServeHTTP(w, r)
	})
}
