package store

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type userRecord struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	LoginID      string     `gorm:"column:login_id;not null;uniqueIndex"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	DisplayName  string     `gorm:"column:display_name;not null"`
	Role         string     `gorm:"column:role;not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (userRecord) TableName() string {
	return "users"
}

type robotRecord struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotCode       string          `gorm:"column:robot_code;not null;uniqueIndex"`
	DisplayName     string          `gorm:"column:display_name;not null"`
	ModelName       *string         `gorm:"column:model_name"`
	Status          string          `gorm:"column:status;not null;default:offline"`
	LastSeenAt      *time.Time      `gorm:"column:last_seen_at"`
	LastStreamingAt *time.Time      `gorm:"column:last_streaming_at"`
	ArchivedAt      *time.Time      `gorm:"column:archived_at"`
	Metadata        json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt       time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;not null;default:now()"`
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
	RobotID        string     `gorm:"column:robot_id;type:uuid;not null;index"`
	TokenHash      string     `gorm:"column:token_hash;not null"`
	TokenPlaintext *string    `gorm:"column:token_plaintext"`
	Name           string     `gorm:"column:name;not null"`
	IsActive       bool       `gorm:"column:is_active;not null;default:true"`
	LastUsedAt     *time.Time `gorm:"column:last_used_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:now()"`
}

func (robotConnectionTokenRecord) TableName() string {
	return "robot_tokens"
}

type missionRecord struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionCode string     `gorm:"column:mission_code;not null;uniqueIndex"`
	Name        string     `gorm:"column:name;not null"`
	MissionType string     `gorm:"column:mission_type;not null"`
	Status      string     `gorm:"column:status;not null;default:ready"`
	CreatedBy   *string    `gorm:"column:created_by;type:uuid"`
	SiteNote    *string    `gorm:"column:site_note"`
	StartedAt   *time.Time `gorm:"column:started_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:now()"`
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
	MissionID string     `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID   string     `gorm:"column:robot_id;type:uuid;not null;index"`
	Role      string     `gorm:"column:role;not null;default:primary"`
	Status    string     `gorm:"column:status;not null;default:assigned"`
	JoinedAt  *time.Time `gorm:"column:joined_at"`
	LeftAt    *time.Time `gorm:"column:left_at"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (missionRobotRecord) TableName() string {
	return "mission_robots"
}

type streamingStatusRecord struct {
	ID                    string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID               string          `gorm:"column:robot_id;type:uuid;not null;uniqueIndex"`
	MissionID             *string         `gorm:"column:mission_id;type:uuid"`
	RoomID                string          `gorm:"column:room_id;not null"`
	Status                string          `gorm:"column:status;not null"`
	PublishedTracks       json.RawMessage `gorm:"column:published_tracks;type:jsonb;not null;default:'[]'::jsonb"`
	PublishedDataChannels json.RawMessage `gorm:"column:published_data_channels;type:jsonb;not null;default:'[]'::jsonb"`
	SentAt                *time.Time      `gorm:"column:sent_at"`
	UpdatedAt             time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (streamingStatusRecord) TableName() string {
	return "streaming_statuses"
}

type telemetrySnapshotRecord struct {
	ID             string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID      string          `gorm:"column:mission_id;type:uuid;not null;index:telemetry_snapshots_mission_received_idx,sort:desc,priority:1"`
	RobotID        string          `gorm:"column:robot_id;type:uuid;not null;index:telemetry_snapshots_robot_received_idx,sort:desc,priority:1"`
	Sequence       *int64          `gorm:"column:sequence"`
	SentAt         *time.Time      `gorm:"column:sent_at"`
	ReceivedAt     time.Time       `gorm:"column:received_at;not null;default:now();index:telemetry_snapshots_mission_received_idx,sort:desc,priority:2;index:telemetry_snapshots_robot_received_idx,sort:desc,priority:2"`
	BatteryPercent *float64        `gorm:"column:battery_percent"`
	NetworkQuality *string         `gorm:"column:network_quality"`
	PositionType   *string         `gorm:"column:position_type"`
	Latitude       *float64        `gorm:"column:latitude"`
	Longitude      *float64        `gorm:"column:longitude"`
	AltitudeMeter  *float64        `gorm:"column:altitude_meter"`
	AccuracyMeter  *float64        `gorm:"column:accuracy_meter"`
	HeadingDegree  *float64        `gorm:"column:heading_degree"`
	Geom           []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RawPayload     json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (telemetrySnapshotRecord) TableName() string {
	return "telemetry_snapshots"
}

type sensorReadingRecord struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID          string          `gorm:"column:mission_id;type:uuid;not null;index:sensor_readings_mission_received_idx,sort:desc,priority:1"`
	RobotID            string          `gorm:"column:robot_id;type:uuid;not null;index:sensor_readings_robot_received_idx,sort:desc,priority:1"`
	Sequence           *int64          `gorm:"column:sequence"`
	SentAt             *time.Time      `gorm:"column:sent_at"`
	ReceivedAt         time.Time       `gorm:"column:received_at;not null;default:now();index:sensor_readings_mission_received_idx,sort:desc,priority:2;index:sensor_readings_robot_received_idx,sort:desc,priority:2"`
	TemperatureCelsius *float64        `gorm:"column:temperature_celsius"`
	HumidityPercent    *float64        `gorm:"column:humidity_percent"`
	OxygenPercent      *float64        `gorm:"column:oxygen_percent"`
	COPpm              *float64        `gorm:"column:co_ppm"`
	CH4Ppm             *float64        `gorm:"column:ch4_ppm"`
	RawPayload         json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (sensorReadingRecord) TableName() string {
	return "sensor_readings"
}

type recordingSessionRecord struct {
	ID                   string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID            string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID              string          `gorm:"column:robot_id;type:uuid;not null;index"`
	RecorderSessionID    *string         `gorm:"column:recorder_session_id;type:uuid"`
	Status               string          `gorm:"column:status;not null;default:pending"`
	ChunkDurationSeconds int             `gorm:"column:chunk_duration_seconds;not null;default:600"`
	StartedAt            time.Time       `gorm:"column:started_at;not null;default:now()"`
	EndedAt              *time.Time      `gorm:"column:ended_at"`
	LastError            *string         `gorm:"column:last_error"`
	Metadata             json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (recordingSessionRecord) TableName() string {
	return "recording_sessions"
}

type recordingChunkRecord struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RecordingSessionID string          `gorm:"column:recording_session_id;type:uuid;not null;uniqueIndex:recording_chunks_session_index_unique,priority:1"`
	MissionID          string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID            string          `gorm:"column:robot_id;type:uuid;not null;index"`
	ChunkIndex         int             `gorm:"column:chunk_index;not null;uniqueIndex:recording_chunks_session_index_unique,priority:2"`
	Status             string          `gorm:"column:status;not null;default:pending"`
	StartedAt          time.Time       `gorm:"column:started_at;not null"`
	EndedAt            *time.Time      `gorm:"column:ended_at"`
	DurationSeconds    *float64        `gorm:"column:duration_seconds"`
	ManifestObjectID   *string         `gorm:"column:manifest_object_id;type:uuid"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (recordingChunkRecord) TableName() string {
	return "recording_chunks"
}

type storageObjectRecord struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID        string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID          *string         `gorm:"column:robot_id;type:uuid"`
	RecordingChunkID *string         `gorm:"column:recording_chunk_id;type:uuid"`
	ObjectType       string          `gorm:"column:object_type;not null"`
	Bucket           string          `gorm:"column:bucket;not null"`
	ObjectKey        string          `gorm:"column:object_key;not null;uniqueIndex"`
	ContentType      *string         `gorm:"column:content_type"`
	SizeBytes        *int64          `gorm:"column:size_bytes"`
	Checksum         *string         `gorm:"column:checksum"`
	StartedAt        *time.Time      `gorm:"column:started_at"`
	EndedAt          *time.Time      `gorm:"column:ended_at"`
	Metadata         json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt        time.Time       `gorm:"column:created_at;not null;default:now()"`
}

func (storageObjectRecord) TableName() string {
	return "storage_objects"
}

type robotSessionRecord struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotID         string          `gorm:"column:robot_id;type:uuid;not null;index"`
	MissionID       *string         `gorm:"column:mission_id;type:uuid"`
	State           string          `gorm:"column:state;not null"`
	ClientIP        *string         `gorm:"column:client_ip;type:inet"`
	UserAgent       *string         `gorm:"column:user_agent"`
	ConnectedAt     time.Time       `gorm:"column:connected_at;not null;default:now()"`
	LastHeartbeatAt time.Time       `gorm:"column:last_heartbeat_at;not null;default:now()"`
	DisconnectedAt  *time.Time      `gorm:"column:disconnected_at"`
	RawPayload      json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (robotSessionRecord) TableName() string {
	return "robot_sessions"
}

type browserSessionRecord struct {
	ID             string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID      string          `gorm:"column:mission_id;type:uuid;not null;index"`
	UserID         *string         `gorm:"column:user_id;type:uuid"`
	State          string          `gorm:"column:state;not null"`
	ConnectedAt    time.Time       `gorm:"column:connected_at;not null;default:now()"`
	DisconnectedAt *time.Time      `gorm:"column:disconnected_at"`
	Metadata       json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (browserSessionRecord) TableName() string {
	return "browser_sessions"
}

type recorderSessionRecord struct {
	ID        string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID string          `gorm:"column:mission_id;type:uuid;not null;index"`
	State     string          `gorm:"column:state;not null"`
	StartedAt time.Time       `gorm:"column:started_at;not null;default:now()"`
	StoppedAt *time.Time      `gorm:"column:stopped_at"`
	LastError *string         `gorm:"column:last_error"`
	Metadata  json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
}

func (recorderSessionRecord) TableName() string {
	return "recorder_sessions"
}

type eventRecord struct {
	ID                     string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID              string          `gorm:"column:mission_id;type:uuid;not null;index:events_mission_occurred_idx,sort:desc,priority:1"`
	RobotID                *string         `gorm:"column:robot_id;type:uuid"`
	EventType              string          `gorm:"column:event_type;not null"`
	Severity               string          `gorm:"column:severity;not null"`
	Title                  string          `gorm:"column:title;not null"`
	Description            *string         `gorm:"column:description"`
	OccurredAt             time.Time       `gorm:"column:occurred_at;not null;index:events_mission_occurred_idx,sort:desc,priority:2"`
	Geom                   []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RelatedStorageObjectID *string         `gorm:"column:related_storage_object_id;type:uuid"`
	RawPayload             json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt              time.Time       `gorm:"column:created_at;not null;default:now()"`
}

func (eventRecord) TableName() string {
	return "events"
}

type controlCommandRecord struct {
	ID            string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID     string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID       string          `gorm:"column:robot_id;type:uuid;not null;index"`
	RequestedBy   *string         `gorm:"column:requested_by;type:uuid"`
	CommandType   string          `gorm:"column:command_type;not null"`
	Status        string          `gorm:"column:status;not null;default:requested"`
	Payload       json.RawMessage `gorm:"column:payload;type:jsonb;not null;default:'{}'::jsonb"`
	RequestedAt   time.Time       `gorm:"column:requested_at;not null;default:now()"`
	SentAt        *time.Time      `gorm:"column:sent_at"`
	CompletedAt   *time.Time      `gorm:"column:completed_at"`
	FailureReason *string         `gorm:"column:failure_reason"`
}

func (controlCommandRecord) TableName() string {
	return "control_commands"
}

type controlAckRecord struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	ControlCommandID string          `gorm:"column:control_command_id;type:uuid;not null;index"`
	RobotID          string          `gorm:"column:robot_id;type:uuid;not null;index"`
	AckStatus        string          `gorm:"column:ack_status;not null"`
	Message          *string         `gorm:"column:message"`
	ReceivedAt       time.Time       `gorm:"column:received_at;not null;default:now()"`
	RawPayload       json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (controlAckRecord) TableName() string {
	return "control_acks"
}

func stringFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
