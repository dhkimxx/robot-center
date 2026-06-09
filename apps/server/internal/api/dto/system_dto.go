package dto

import (
	"robot-center/apps/server/internal/service"
	"robot-center/apps/server/internal/sfu"
	"robot-center/apps/server/internal/store"
)

type SystemStatusResponse struct {
	Service       string                      `json:"service"`
	Status        string                      `json:"status"`
	Components    []SystemComponentStatus     `json:"components"`
	Config        SystemConfigResponse        `json:"config"`
	ObjectStorage ObjectStorageStatusResponse `json:"objectStorage"`
	Database      DatabaseStatusResponse      `json:"database"`
	Summary       SystemSummaryResponse       `json:"summary"`
	SFURooms      []sfu.RoomSummary           `json:"sfuRooms"`
}

type SystemStatusInput struct {
	Environment               string
	AppServerPublicURL        string
	RecorderWorkerInternalURL string
	MinIOInternalURL          string
	MinIOPublicURL            string
	MinIOBucket               string
	RecorderWorkerStatus      string
	ObjectStorage             ObjectStorageStatusResponse
	Database                  DatabaseStatusResponse
	RobotCount                int
	MissionCount              int
	RecordingCount            int
	SFURooms                  []sfu.RoomSummary
}

type SystemComponentStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type SystemConfigResponse struct {
	Environment               string `json:"environment"`
	AppServerPublicURL        string `json:"appServerPublicUrl"`
	RecorderWorkerInternalURL string `json:"recorderWorkerInternalUrl"`
	MinIOInternalURL          string `json:"minioInternalUrl"`
	MinIOPublicURL            string `json:"minioPublicUrl,omitempty"`
	MinIOBucket               string `json:"minioBucket"`
}

type SystemSummaryResponse struct {
	Robots     int `json:"robots"`
	Missions   int `json:"missions"`
	SFURooms   int `json:"sfuRooms"`
	Recordings int `json:"recordings"`
}

type ObjectStorageStatusResponse struct {
	Status             string   `json:"status"`
	Bucket             string   `json:"bucket"`
	ObjectCount        *int     `json:"objectCount,omitempty"`
	BucketUsedBytes    *int64   `json:"bucketUsedBytes,omitempty"`
	TotalBytes         *int64   `json:"totalBytes,omitempty"`
	UsedBytes          *int64   `json:"usedBytes,omitempty"`
	AvailableBytes     *int64   `json:"availableBytes,omitempty"`
	UsedPercent        *float64 `json:"usedPercent,omitempty"`
	DiskTotalBytes     *int64   `json:"diskTotalBytes,omitempty"`
	DiskUsedBytes      *int64   `json:"diskUsedBytes,omitempty"`
	DiskAvailableBytes *int64   `json:"diskAvailableBytes,omitempty"`
	DiskUsedPercent    *float64 `json:"diskUsedPercent,omitempty"`
	Error              string   `json:"error,omitempty"`
}

type DatabaseStatusResponse struct {
	Status            string                       `json:"status"`
	DatabaseName      string                       `json:"databaseName,omitempty"`
	DatabaseSizeBytes *int64                       `json:"databaseSizeBytes,omitempty"`
	TrackedTableBytes *int64                       `json:"trackedTableBytes,omitempty"`
	Tables            []DatabaseTableUsageResponse `json:"tables,omitempty"`
	Error             string                       `json:"error,omitempty"`
}

type DatabaseTableUsageResponse struct {
	TableName  string `json:"tableName"`
	RowCount   int64  `json:"rowCount"`
	TotalBytes int64  `json:"totalBytes"`
}

type ClearObjectStorageRequest struct {
	Confirmation string `json:"confirmation"`
}

type ClearObjectStorageResponse struct {
	ObjectStorage service.ObjectStorageClearResult `json:"objectStorage"`
}

type ClearSensorDataRequest struct {
	Confirmation string `json:"confirmation"`
}

type ClearSensorDataResponse struct {
	SensorData store.SensorDataClearResult `json:"sensorData"`
}

type ClearEventDataRequest struct {
	Confirmation string `json:"confirmation"`
}

type ClearEventDataResponse struct {
	EventData store.EventDataClearResult `json:"eventData"`
}

func SystemStatus(input SystemStatusInput) SystemStatusResponse {
	return SystemStatusResponse{
		Service: "app-server",
		Status:  "ok",
		Components: []SystemComponentStatus{
			{Name: "app-server", Status: "ok"},
			{Name: "recorder-worker", Status: input.RecorderWorkerStatus},
			{Name: "turn", Status: "configured"},
			{Name: "postgres", Status: "configured"},
			{Name: "minio", Status: "configured"},
		},
		Config: SystemConfigResponse{
			Environment:               input.Environment,
			AppServerPublicURL:        input.AppServerPublicURL,
			RecorderWorkerInternalURL: input.RecorderWorkerInternalURL,
			MinIOInternalURL:          input.MinIOInternalURL,
			MinIOPublicURL:            input.MinIOPublicURL,
			MinIOBucket:               input.MinIOBucket,
		},
		ObjectStorage: input.ObjectStorage,
		Database:      input.Database,
		Summary: SystemSummaryResponse{
			Robots:     input.RobotCount,
			Missions:   input.MissionCount,
			SFURooms:   len(input.SFURooms),
			Recordings: input.RecordingCount,
		},
		SFURooms: input.SFURooms,
	}
}

func DatabaseStatus(usage store.DatabaseUsageResult) DatabaseStatusResponse {
	return DatabaseStatusResponse{
		Status:            usage.Status,
		DatabaseName:      usage.DatabaseName,
		DatabaseSizeBytes: int64Ptr(usage.DatabaseSizeBytes),
		TrackedTableBytes: int64Ptr(usage.TrackedTableBytes),
		Tables:            DatabaseTableUsage(usage.Tables),
	}
}

func DatabaseUnavailable(err error) DatabaseStatusResponse {
	response := DatabaseStatusResponse{Status: "unavailable"}
	if err != nil {
		response.Error = err.Error()
	}
	return response
}

func DatabaseTableUsage(tables []store.DatabaseTableUsage) []DatabaseTableUsageResponse {
	response := make([]DatabaseTableUsageResponse, 0, len(tables))
	for _, table := range tables {
		response = append(response, DatabaseTableUsageResponse{
			TableName:  table.TableName,
			RowCount:   table.RowCount,
			TotalBytes: table.TotalBytes,
		})
	}
	return response
}

func ObjectStorageStatus(usage service.ObjectStorageUsageResult) ObjectStorageStatusResponse {
	return ObjectStorageStatusResponse{
		Status:             usage.Status,
		Bucket:             usage.Bucket,
		ObjectCount:        intPtr(usage.ObjectCount),
		BucketUsedBytes:    int64Ptr(usage.BucketUsedBytes),
		TotalBytes:         int64Ptr(usage.TotalBytes),
		UsedBytes:          int64Ptr(usage.UsedBytes),
		AvailableBytes:     int64Ptr(usage.AvailableBytes),
		UsedPercent:        float64Ptr(usage.UsedPercent),
		DiskTotalBytes:     int64Ptr(usage.DiskTotalBytes),
		DiskUsedBytes:      int64Ptr(usage.DiskUsedBytes),
		DiskAvailableBytes: int64Ptr(usage.DiskAvailableBytes),
		DiskUsedPercent:    float64Ptr(usage.DiskUsedPercent),
	}
}

func ObjectStorageUnavailable(bucket string, err error) ObjectStorageStatusResponse {
	response := ObjectStorageStatusResponse{
		Status: "unavailable",
		Bucket: bucket,
	}
	if err != nil {
		response.Error = err.Error()
	}
	return response
}

func intPtr(value int) *int {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

func float64Ptr(value float64) *float64 {
	return &value
}
