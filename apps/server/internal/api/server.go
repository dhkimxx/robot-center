package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/service"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"

	"github.com/gin-gonic/gin"
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
		Endpoint:    cfg.MinIOInternalURL,
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
		ValidateRobotSelection: func(roomID string, robotCode string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return services.Missions.ValidateActiveMissionRobot(ctx, roomID, robotCode)
		},
		OnPublisherEvent: func(event sfu.PublisherEvent) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := services.Streams.HandlePublisherEvent(ctx, event); err != nil {
				log.Printf("stream session publisher event failed type=%s room=%s robot=%s peer=%s: %v", event.Type, event.RoomID, event.RobotCode, event.PublisherPeerID, err)
			}
		},
	})
	return server
}

func NewServerFromConfig(ctx context.Context, cfg config.AppServerConfig) (*Server, error) {
	repository, err := store.NewPostgresStore(ctx, store.PostgresConfig{
		DSN:                cfg.PostgresDSN,
		AppServerPublicURL: cfg.AppServerPublicURL,
		MinIOBucket:        cfg.MinIOBucket,
	})
	if err != nil {
		return nil, err
	}
	return NewServerWithStore(cfg, repository), nil
}

func (s *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.HandleMethodNotAllowed = true
	router.Use(gin.Recovery(), appHeaderMiddleware())

	s.registerRoutes(router)

	if s.config.WebStaticDir != "" {
		router.NoRoute(gin.WrapH(s.staticHandler()))
	}

	return router
}
