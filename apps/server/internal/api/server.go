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
	return &Server{
		config:   cfg,
		services: service.NewServices(repository),
		sfuHub: sfu.NewHub(sfu.Config{
			TURNURL:      cfg.TURNURL,
			TURNUsername: cfg.TURNUsername,
			TURNPassword: cfg.TURNPassword,
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
	mux.HandleFunc("GET /api/telemetry", s.handleListTelemetry)
	mux.HandleFunc("POST /api/telemetry", s.handleCreateTelemetry)
	mux.HandleFunc("GET /api/sensor-readings", s.handleListSensorReadings)
	mux.HandleFunc("POST /api/sensor-readings", s.handleCreateSensorReading)
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

func (s *Server) handleListTelemetry(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	telemetry, err := s.services.Telemetry.ListTelemetry(r.Context(), missionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"telemetry": dto.TelemetryList(telemetry),
	})
}

func (s *Server) handleCreateTelemetry(w http.ResponseWriter, r *http.Request) {
	envelope, rawPayload, err := decodeDataChannelEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	position := mapValue(envelope.Payload, "position")
	snapshot := domain.TelemetrySnapshot{
		RobotCode:           envelope.RobotCode,
		MissionID:           envelope.MissionID,
		MessageID:           envelope.MessageID,
		MessageType:         envelope.MessageType,
		Sequence:            envelope.Sequence,
		SentAt:              envelope.SentAt,
		BatteryPercent:      floatPointer(envelope.Payload, "batteryPercent"),
		NetworkState:        stringValue(envelope.Payload, "networkState"),
		PositionAvailable:   boolValue(envelope.Payload, "positionAvailable"),
		Latitude:            floatPointer(position, "latitude"),
		Longitude:           floatPointer(position, "longitude"),
		AltitudeMeter:       floatPointer(position, "altitudeMeter"),
		AccuracyMeter:       floatPointer(position, "accuracyMeter"),
		HeadingDegree:       floatPointer(position, "headingDegree"),
		SpeedMeterPerSecond: floatPointer(position, "speedMeterPerSecond"),
		RawPayload:          rawPayload,
	}
	snapshot, err = s.services.Telemetry.SaveTelemetry(r.Context(), snapshot)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"telemetry": dto.Telemetry(snapshot),
	})
}

func (s *Server) handleListSensorReadings(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	readings, err := s.services.Sensors.ListSensorReadings(r.Context(), missionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sensorReadings": dto.SensorReadings(readings),
	})
}

func (s *Server) handleCreateSensorReading(w http.ResponseWriter, r *http.Request) {
	envelope, rawPayload, err := decodeDataChannelEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	reading := domain.SensorReading{
		RobotCode:          envelope.RobotCode,
		MissionID:          envelope.MissionID,
		MessageID:          envelope.MessageID,
		MessageType:        envelope.MessageType,
		Sequence:           envelope.Sequence,
		SentAt:             envelope.SentAt,
		BatteryPercent:     floatPointer(envelope.Payload, "batteryPercent"),
		TemperatureCelsius: floatPointer(envelope.Payload, "temperatureCelsius"),
		HumidityPercent:    floatPointer(envelope.Payload, "humidityPercent"),
		OxygenPercent:      floatPointer(envelope.Payload, "oxygenPercent"),
		COPpm:              floatPointer(envelope.Payload, "coPpm"),
		CH4Ppm:             floatPointer(envelope.Payload, "ch4Ppm"),
		RawPayload:         rawPayload,
	}
	reading, err = s.services.Sensors.SaveSensorReading(r.Context(), reading)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"sensorReading": dto.SensorReading(reading),
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
		writeError(w, http.StatusBadRequest, err)
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

type dataChannelEnvelope struct {
	MessageID   string         `json:"messageId"`
	MessageType string         `json:"messageType"`
	RobotCode   string         `json:"robotCode"`
	MissionID   string         `json:"missionId"`
	Sequence    int64          `json:"sequence"`
	SentAt      *time.Time     `json:"sentAt"`
	Payload     map[string]any `json:"payload"`
}

func decodeDataChannelEnvelope(r *http.Request) (dataChannelEnvelope, json.RawMessage, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return dataChannelEnvelope{}, nil, err
	}

	var envelope dataChannelEnvelope
	if err := json.Unmarshal(rawPayload, &envelope); err != nil {
		return dataChannelEnvelope{}, nil, err
	}
	if envelope.Payload == nil {
		envelope.Payload = map[string]any{}
	}
	if strings.TrimSpace(envelope.RobotCode) == "" {
		return dataChannelEnvelope{}, nil, errors.New("robotCode is required")
	}
	if strings.TrimSpace(envelope.MissionID) == "" {
		return dataChannelEnvelope{}, nil, errors.New("missionId is required")
	}
	envelope.RobotCode = strings.TrimSpace(envelope.RobotCode)
	envelope.MissionID = strings.TrimSpace(envelope.MissionID)
	return envelope, append(json.RawMessage(nil), rawPayload...), nil
}

func mapValue(values map[string]any, key string) map[string]any {
	value, ok := values[key]
	if !ok {
		return map[string]any{}
	}
	asMap, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return asMap
}

func stringValue(values map[string]any, key string) string {
	value, ok := values[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func boolValue(values map[string]any, key string) bool {
	value, ok := values[key].(bool)
	return ok && value
}

func floatPointer(values map[string]any, key string) *float64 {
	value, ok := values[key]
	if !ok {
		return nil
	}
	switch typedValue := value.(type) {
	case float64:
		return &typedValue
	case int:
		converted := float64(typedValue)
		return &converted
	default:
		return nil
	}
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
