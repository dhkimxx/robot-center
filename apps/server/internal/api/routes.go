package api

import "github.com/gin-gonic/gin"

func (s *Server) registerRoutes(router *gin.Engine) {
	router.GET("/healthz", ginHTTPHandler(s.handleHealth))
	router.GET("/swagger/index.html", ginHTTPHandler(s.handleSwaggerUI))
	router.GET("/swagger/doc.json", ginHTTPHandler(s.handleSwaggerDocJSON))
	router.GET("/swagger/openapi.json", ginHTTPHandler(s.handleOpenAPIJSON))

	v1 := router.Group("/api/v1")
	s.registerSystemRoutes(v1.Group("/system", s.systemAuthMiddleware()))
	s.registerOperatorRoutes(v1.Group("/operator", s.operatorAuthMiddleware()))
	s.registerRecorderRoutes(v1.Group("/recorder", s.recorderAuthMiddleware()))
	s.registerRobotRoutes(v1.Group("/robot", s.robotAuthMiddleware()))
}

func (s *Server) registerSystemRoutes(group *gin.RouterGroup) {
	group.GET("/status", ginHTTPHandler(s.handleSystemStatus))
	group.POST("/object-storage/clear", ginHTTPHandler(s.handleClearObjectStorage))
	group.POST("/object-storage/prune", ginHTTPHandler(s.handlePruneObjectStorage))
	group.POST("/sensors/clear", ginHTTPHandler(s.handleClearSensorData))
	group.POST("/events/clear", ginHTTPHandler(s.handleClearEventData))
	group.POST("/recorder-runtime/clear", ginHTTPHandler(s.handleClearRecorderRuntime))
	group.POST("/recorder-runtime/prune", ginHTTPHandler(s.handlePruneRecorderRuntime))
}

func (s *Server) registerOperatorRoutes(group *gin.RouterGroup) {
	group.GET("/rtc-config", ginHTTPHandler(s.handleRTCConfig))
	group.GET("/sensor-descriptors", ginHTTPHandler(s.handleListSensorDescriptors))
	group.GET("/sensor-samples", ginHTTPHandler(s.handleListSensorSamples))
	group.GET("/sensor-latest", ginHTTPHandler(s.handleListSensorLatest))
	group.GET("/recordings", ginHTTPHandler(s.handleListRecordings))

	group.GET("/robots", ginHTTPHandler(s.handleListRobots))
	group.POST("/robots", ginHTTPHandler(s.handleCreateRobot))
	group.PATCH("/robots/:robotCode", ginHTTPHandler(s.handleUpdateRobot, "robotCode"))
	group.DELETE("/robots/:robotCode", ginHTTPHandler(s.handleArchiveRobot, "robotCode"))
	group.GET("/robots/:robotCode/connection-info", ginHTTPHandler(s.handleGetRobotConnectionInfo, "robotCode"))
	group.POST("/robots/:robotCode/connection-token", ginHTTPHandler(s.handleRotateRobotConnectionToken, "robotCode"))

	group.GET("/missions", ginHTTPHandler(s.handleListMissions))
	group.POST("/missions", ginHTTPHandler(s.handleCreateMission))
	group.GET("/missions/:missionCode/events", ginHTTPHandler(s.handleListMissionEvents, "missionCode"))
	group.GET("/missions/:missionCode/live-status", ginHTTPHandler(s.handleMissionLiveStatus, "missionCode"))
	group.GET("/missions/:missionCode/recordings/summary", ginHTTPHandler(s.handleMissionRecordingSummary, "missionCode"))
	group.GET("/missions/:missionCode/recordings/chunks", ginHTTPHandler(s.handleMissionRecordingChunks, "missionCode"))
	group.POST("/missions/:missionCode/start", ginHTTPHandler(s.handleStartMission, "missionCode"))
	group.POST("/missions/:missionCode/end", ginHTTPHandler(s.handleEndMission, "missionCode"))

	group.GET("/sfu/ws", ginHTTPHandler(s.handleOperatorSFUWebSocket))
}

func (s *Server) registerRecorderRoutes(group *gin.RouterGroup) {
	group.GET("/recording-targets", ginHTTPHandler(s.handleRecordingTargets))
	group.POST("/tick", ginHTTPHandler(s.handleRecorderTick))
	group.POST("/finalization-jobs/claim", ginHTTPHandler(s.handleRecorderFinalizationJobsClaim))
	group.POST("/finalization-jobs/:jobID/completed", ginHTTPHandler(s.handleRecorderFinalizationJobCompleted, "jobID"))
	group.POST("/finalization-jobs/:jobID/partial", ginHTTPHandler(s.handleRecorderFinalizationJobPartial, "jobID"))
	group.POST("/finalization-jobs/:jobID/failed", ginHTTPHandler(s.handleRecorderFinalizationJobFailed, "jobID"))
	group.POST("/chunks/:chunkID/uploaded", ginHTTPHandler(s.handleRecorderChunkUploaded, "chunkID"))
	group.POST("/chunks/:chunkID/files/:fileType/uploaded", ginHTTPHandler(s.handleRecorderFileUploaded, "chunkID", "fileType"))
	group.POST("/sensor-samples", ginHTTPHandler(s.handleCreateSensorSamples))
	group.POST("/events", ginHTTPHandler(s.handleCreateMissionEvents))
	group.GET("/sfu/ws", ginHTTPHandler(s.handleRecorderSFUWebSocket))
}

func (s *Server) registerRobotRoutes(group *gin.RouterGroup) {
	group.POST("/heartbeat", ginHTTPHandler(s.handleRobotAPIHeartbeat))
	group.GET("/mission", ginHTTPHandler(s.handleRobotAPIMission))
	group.GET("/sfu/ws", ginHTTPHandler(s.handleRobotSFUWebSocket))
}
