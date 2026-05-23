package store

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type robotRecord struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotCode       string          `gorm:"column:robot_code"`
	DisplayName     string          `gorm:"column:display_name"`
	ModelName       *string         `gorm:"column:model_name"`
	Status          string          `gorm:"column:status"`
	LastSeenAt      *time.Time      `gorm:"column:last_seen_at"`
	LastStreamingAt *time.Time      `gorm:"column:last_streaming_at"`
	ArchivedAt      *time.Time      `gorm:"column:archived_at"`
	Metadata        json.RawMessage `gorm:"column:metadata;type:jsonb;default:'{}'"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
}

func (robotRecord) TableName() string {
	return "robots"
}

func (record robotRecord) toDomainRobot() domain.Robot {
	return domain.Robot{
		ID:              record.ID,
		RobotCode:       record.RobotCode,
		DisplayName:     record.DisplayName,
		ModelName:       stringFromPointer(record.ModelName),
		Status:          record.Status,
		LastSeenAt:      record.LastSeenAt,
		LastStreamingAt: record.LastStreamingAt,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

type robotConnectionTokenRecord struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID        string     `gorm:"column:robot_id;type:uuid"`
	TokenHash      string     `gorm:"column:token_hash"`
	TokenPlaintext *string    `gorm:"column:token_plaintext"`
	Name           string     `gorm:"column:name"`
	IsActive       bool       `gorm:"column:is_active;default:true"`
	LastUsedAt     *time.Time `gorm:"column:last_used_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
}

func (robotConnectionTokenRecord) TableName() string {
	return "robot_tokens"
}

type missionRecord struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionCode string     `gorm:"column:mission_code"`
	Name        string     `gorm:"column:name"`
	MissionType string     `gorm:"column:mission_type"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *string    `gorm:"column:created_by;type:uuid"`
	SiteNote    *string    `gorm:"column:site_note"`
	StartedAt   *time.Time `gorm:"column:started_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (missionRecord) TableName() string {
	return "missions"
}

func (record missionRecord) toDomainMission(robotCodes []string) domain.Mission {
	robotCodes = append([]string(nil), robotCodes...)
	return domain.Mission{
		ID:          record.ID,
		MissionCode: record.MissionCode,
		Name:        record.Name,
		MissionType: record.MissionType,
		Status:      record.Status,
		SiteNote:    stringFromPointer(record.SiteNote),
		RobotCode:   firstString(robotCodes),
		RobotCodes:  robotCodes,
		StartedAt:   record.StartedAt,
		EndedAt:     record.EndedAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

type missionRobotRecord struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID string     `gorm:"column:mission_id;type:uuid"`
	RobotID   string     `gorm:"column:robot_id;type:uuid"`
	Role      string     `gorm:"column:role;default:primary"`
	Status    string     `gorm:"column:status;default:assigned"`
	JoinedAt  *time.Time `gorm:"column:joined_at"`
	LeftAt    *time.Time `gorm:"column:left_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (missionRobotRecord) TableName() string {
	return "mission_robots"
}

type streamingStatusRecord struct {
	ID                    string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID               string          `gorm:"column:robot_id;type:uuid"`
	MissionID             *string         `gorm:"column:mission_id;type:uuid"`
	RoomID                string          `gorm:"column:room_id"`
	Status                string          `gorm:"column:status"`
	PublishedTracks       json.RawMessage `gorm:"column:published_tracks;type:jsonb;default:'[]'"`
	PublishedDataChannels json.RawMessage `gorm:"column:published_data_channels;type:jsonb;default:'[]'"`
	SentAt                *time.Time      `gorm:"column:sent_at"`
	UpdatedAt             time.Time       `gorm:"column:updated_at"`
}

func (streamingStatusRecord) TableName() string {
	return "streaming_statuses"
}

type telemetrySnapshotRecord struct {
	ID             string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID      string          `gorm:"column:mission_id;type:uuid"`
	RobotID        string          `gorm:"column:robot_id;type:uuid"`
	Sequence       *int64          `gorm:"column:sequence"`
	SentAt         *time.Time      `gorm:"column:sent_at"`
	ReceivedAt     time.Time       `gorm:"column:received_at"`
	BatteryPercent *float64        `gorm:"column:battery_percent"`
	NetworkQuality *string         `gorm:"column:network_quality"`
	PositionType   *string         `gorm:"column:position_type"`
	Latitude       *float64        `gorm:"column:latitude"`
	Longitude      *float64        `gorm:"column:longitude"`
	AltitudeMeter  *float64        `gorm:"column:altitude_meter"`
	AccuracyMeter  *float64        `gorm:"column:accuracy_meter"`
	HeadingDegree  *float64        `gorm:"column:heading_degree"`
	Geom           []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RawPayload     json.RawMessage `gorm:"column:raw_payload;type:jsonb;default:'{}'"`
}

func (telemetrySnapshotRecord) TableName() string {
	return "telemetry_snapshots"
}

type sensorReadingRecord struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID          string          `gorm:"column:mission_id;type:uuid"`
	RobotID            string          `gorm:"column:robot_id;type:uuid"`
	Sequence           *int64          `gorm:"column:sequence"`
	SentAt             *time.Time      `gorm:"column:sent_at"`
	ReceivedAt         time.Time       `gorm:"column:received_at"`
	TemperatureCelsius *float64        `gorm:"column:temperature_celsius"`
	HumidityPercent    *float64        `gorm:"column:humidity_percent"`
	OxygenPercent      *float64        `gorm:"column:oxygen_percent"`
	COPpm              *float64        `gorm:"column:co_ppm"`
	CH4Ppm             *float64        `gorm:"column:ch4_ppm"`
	RawPayload         json.RawMessage `gorm:"column:raw_payload;type:jsonb;default:'{}'"`
}

func (sensorReadingRecord) TableName() string {
	return "sensor_readings"
}

type recordingChunkRecord struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RecordingSessionID string          `gorm:"column:recording_session_id;type:uuid"`
	MissionID          string          `gorm:"column:mission_id;type:uuid"`
	RobotID            string          `gorm:"column:robot_id;type:uuid"`
	ChunkIndex         int             `gorm:"column:chunk_index"`
	Status             string          `gorm:"column:status"`
	StartedAt          time.Time       `gorm:"column:started_at"`
	EndedAt            *time.Time      `gorm:"column:ended_at"`
	DurationSeconds    *float64        `gorm:"column:duration_seconds"`
	ManifestObjectID   *string         `gorm:"column:manifest_object_id;type:uuid"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;default:'{}'"`
	CreatedAt          time.Time       `gorm:"column:created_at"`
	UpdatedAt          time.Time       `gorm:"column:updated_at"`
}

func (recordingChunkRecord) TableName() string {
	return "recording_chunks"
}

type storageObjectRecord struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID        string          `gorm:"column:mission_id;type:uuid"`
	RobotID          *string         `gorm:"column:robot_id;type:uuid"`
	RecordingChunkID *string         `gorm:"column:recording_chunk_id;type:uuid"`
	ObjectType       string          `gorm:"column:object_type"`
	Bucket           string          `gorm:"column:bucket"`
	ObjectKey        string          `gorm:"column:object_key"`
	ContentType      *string         `gorm:"column:content_type"`
	SizeBytes        *int64          `gorm:"column:size_bytes"`
	Checksum         *string         `gorm:"column:checksum"`
	StartedAt        *time.Time      `gorm:"column:started_at"`
	EndedAt          *time.Time      `gorm:"column:ended_at"`
	Metadata         json.RawMessage `gorm:"column:metadata;type:jsonb;default:'{}'"`
	CreatedAt        time.Time       `gorm:"column:created_at"`
}

func (storageObjectRecord) TableName() string {
	return "storage_objects"
}

func stringFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
