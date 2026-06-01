package api

import (
	"context"
	"net/http"
	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/service"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
	"time"
)

type Server struct {
	config   config.AppServerConfig
	services *service.Services
	sfuHub   *sfu.Hub
	started  time.Time
}

func NewServerWithStore(cfg config.AppServerConfig, repository store.Store) *Server {
	services := service.NewServices(repository)
	services.Storage = service.NewObjectStorageAdminService(service.ObjectStorageAdminConfig{
		Environment: cfg.Environment,
		Endpoint:    cfg.MinIOEndpoint,
		Bucket:      cfg.MinIOBucket,
		AccessKey:   cfg.MinIOAccessKey,
		SecretKey:   cfg.MinIOSecretKey,
	}, repository)
	server := &Server{
		config:   cfg,
		services: services,
		started:  time.Now().UTC(),
	}
	server.sfuHub = sfu.NewHub(sfu.Config{
		TURNURL:      cfg.TURNInternalURL,
		TURNUsername: cfg.TURNUsername,
		TURNPassword: cfg.TURNPassword,
		ValidateRobotPublisher: func(roomID string, robotCode string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return services.Missions.ValidateActiveMissionRobot(ctx, roomID, robotCode)
		},
	})
	return server
}

func NewServerFromConfig(ctx context.Context, cfg config.AppServerConfig) (*Server, error) {
	repository, err := store.NewPostgresStore(ctx, store.PostgresConfig{
		DSN:         cfg.PostgresDSN,
		ServerURL:   cfg.PublicURL,
		MinIOBucket: cfg.MinIOBucket,
	})
	if err != nil {
		return nil, err
	}
	return NewServerWithStore(cfg, repository), nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/docs", s.handleSwaggerUI)
	mux.HandleFunc("GET /api/docs/", s.handleSwaggerUI)
	mux.HandleFunc("GET /api/docs/openapi.json", s.handleOpenAPIJSON)
	mux.HandleFunc("GET /api/system/status", s.handleSystemStatus)
	mux.HandleFunc("POST /api/system/object-storage/clear", s.handleClearObjectStorage)
	mux.HandleFunc("GET /api/rtc-config", s.handleRTCConfig)
	mux.HandleFunc("GET /api/recording-targets", s.handleRecordingTargets)
	mux.HandleFunc("GET /api/sensor-descriptors", s.handleListSensorDescriptors)
	mux.HandleFunc("POST /api/sensor-descriptors", s.handleCreateSensorSamples)
	mux.HandleFunc("GET /api/sensor-samples", s.handleListSensorSamples)
	mux.HandleFunc("POST /api/sensor-samples", s.handleCreateSensorSamples)
	mux.HandleFunc("GET /api/sensor-latest", s.handleListSensorLatest)
	mux.HandleFunc("GET /api/recordings", s.handleListRecordings)
	mux.HandleFunc("POST /api/recorder/tick", s.handleRecorderTick)
	mux.HandleFunc("POST /api/recorder/finalization-jobs/claim", s.handleRecorderFinalizationJobsClaim)
	mux.HandleFunc("POST /api/recorder/finalization-jobs/{jobID}/completed", s.handleRecorderFinalizationJobCompleted)
	mux.HandleFunc("POST /api/recorder/finalization-jobs/{jobID}/partial", s.handleRecorderFinalizationJobPartial)
	mux.HandleFunc("POST /api/recorder/finalization-jobs/{jobID}/failed", s.handleRecorderFinalizationJobFailed)
	mux.HandleFunc("POST /api/recorder/chunks/{chunkID}/uploaded", s.handleRecorderChunkUploaded)
	mux.HandleFunc("POST /api/recorder/chunks/{chunkID}/files/{fileType}/uploaded", s.handleRecorderFileUploaded)
	mux.HandleFunc("GET /api/robots", s.handleListRobots)
	mux.HandleFunc("POST /api/robots", s.handleCreateRobot)
	mux.HandleFunc("PATCH /api/robots/{robotCode}", s.handleUpdateRobot)
	mux.HandleFunc("DELETE /api/robots/{robotCode}", s.handleArchiveRobot)
	mux.HandleFunc("GET /api/robots/{robotCode}/connection-info", s.handleGetRobotConnectionInfo)
	mux.HandleFunc("POST /api/robots/{robotCode}/connection-token", s.handleRotateRobotConnectionToken)
	mux.HandleFunc("GET /api/missions", s.handleListMissions)
	mux.HandleFunc("GET /api/missions/{missionCode}/live-status", s.handleMissionLiveStatus)
	mux.HandleFunc("POST /api/missions", s.handleCreateMission)
	mux.HandleFunc("POST /api/missions/{missionCode}/start", s.handleStartMission)
	mux.HandleFunc("POST /api/missions/{missionCode}/end", s.handleEndMission)
	mux.HandleFunc("POST /api/v1/robot/heartbeat", s.handleRobotAPIHeartbeat)
	mux.HandleFunc("GET /api/v1/robot/mission", s.handleRobotAPIMission)
	mux.HandleFunc("GET /api/v1/robot/sfu/ws", s.handleRobotSFUWebSocket)
	mux.HandleFunc("GET /sfu/operator/ws", s.handleOperatorSFUWebSocket)
	mux.HandleFunc("GET /sfu/recorder/ws", s.handleRecorderSFUWebSocket)

	if s.config.WebStaticDir != "" {
		mux.Handle("/", s.staticHandler())
	}

	return withRequestHeaders(mux)
}
