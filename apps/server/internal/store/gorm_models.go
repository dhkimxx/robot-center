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
	ID          string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RobotCode   string          `gorm:"column:robot_code;not null;uniqueIndex"`
	DisplayName string          `gorm:"column:display_name;not null"`
	ModelName   *string         `gorm:"column:model_name"`
	Status      string          `gorm:"column:status;not null;default:offline"`
	LastSeenAt  *time.Time      `gorm:"column:last_seen_at"`
	ArchivedAt  *time.Time      `gorm:"column:archived_at"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt   time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (robotRecord) TableName() string {
	return "robots"
}

func (record robotRecord) toDomainRobot() domain.Robot {
	return domain.Robot{
		ID:          record.ID,
		RobotCode:   record.RobotCode,
		DisplayName: record.DisplayName,
		ModelName:   stringFromPointer(record.ModelName),
		Status:      record.Status,
		LastSeenAt:  record.LastSeenAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
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
	CreatedBy   *string    `gorm:"column:created_by;type:uuid;index"`
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

type sensorDescriptorRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID    string          `gorm:"column:mission_id;type:uuid;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:1;index"`
	RobotID      string          `gorm:"column:robot_id;type:uuid;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:2;index"`
	SensorID     string          `gorm:"column:sensor_id;not null;uniqueIndex:sensor_descriptors_mission_robot_sensor_unique,priority:3"`
	ChannelRole  string          `gorm:"column:channel_role;not null"`
	DisplayName  string          `gorm:"column:display_name;not null"`
	SensorType   string          `gorm:"column:sensor_type;not null;index"`
	ValueType    string          `gorm:"column:value_type;not null"`
	Unit         *string         `gorm:"column:unit"`
	SampleRateHz *float64        `gorm:"column:sample_rate_hz"`
	Enabled      bool            `gorm:"column:enabled;not null"`
	Metadata     json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	FirstSeenAt  time.Time       `gorm:"column:first_seen_at;not null;default:now()"`
	LastSeenAt   time.Time       `gorm:"column:last_seen_at;not null;default:now();index"`
}

func (sensorDescriptorRecord) TableName() string {
	return "sensor_descriptors"
}

type sensorSampleRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	DescriptorID string          `gorm:"column:descriptor_id;type:uuid;not null;index;index:sensor_samples_descriptor_received_idx,priority:1"`
	MissionID    string          `gorm:"column:mission_id;type:uuid;not null;index:sensor_samples_latest_idx,priority:1"`
	RobotID      string          `gorm:"column:robot_id;type:uuid;not null;index:sensor_samples_latest_idx,priority:2"`
	SensorID     string          `gorm:"column:sensor_id;not null;index:sensor_samples_latest_idx,priority:3"`
	ChannelRole  string          `gorm:"column:channel_role;not null;index"`
	MessageID    *string         `gorm:"column:message_id;index"`
	Sequence     *int64          `gorm:"column:sequence"`
	SentAt       *time.Time      `gorm:"column:sent_at"`
	ReceivedAt   time.Time       `gorm:"column:received_at;not null;default:now();index:sensor_samples_latest_idx,sort:desc,priority:4;index:sensor_samples_descriptor_received_idx,sort:desc,priority:2"`
	NumericValue *float64        `gorm:"column:numeric_value"`
	TextValue    *string         `gorm:"column:text_value"`
	BoolValue    *bool           `gorm:"column:bool_value"`
	VectorValue  json.RawMessage `gorm:"column:vector_value;type:jsonb"`
	ObjectValue  json.RawMessage `gorm:"column:object_value;type:jsonb"`
	ObjectKey    *string         `gorm:"column:object_key"`
	RawPayload   json.RawMessage `gorm:"column:raw_payload;type:jsonb;not null;default:'{}'::jsonb"`
}

func (sensorSampleRecord) TableName() string {
	return "sensor_samples"
}

type recordingSessionRecord struct {
	ID                   string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID            string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID              string          `gorm:"column:robot_id;type:uuid;not null;index"`
	RecorderSessionID    *string         `gorm:"column:recorder_session_id;type:uuid;index"`
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
	ManifestObjectID   *string         `gorm:"column:manifest_object_id;type:uuid;index"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (recordingChunkRecord) TableName() string {
	return "recording_chunks"
}

type recordingFinalizationJobRecord struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RecordingChunkID   string          `gorm:"column:recording_chunk_id;type:uuid;not null;uniqueIndex"`
	RecordingSessionID string          `gorm:"column:recording_session_id;type:uuid;not null;index"`
	MissionID          string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID            string          `gorm:"column:robot_id;type:uuid;not null;index"`
	Status             string          `gorm:"column:status;not null;default:queued;index"`
	Reason             *string         `gorm:"column:reason"`
	Attempts           int             `gorm:"column:attempts;not null;default:0"`
	LockedBy           *string         `gorm:"column:locked_by"`
	LockedUntil        *time.Time      `gorm:"column:locked_until;index"`
	LastError          *string         `gorm:"column:last_error"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'::jsonb"`
	CompletedAt        *time.Time      `gorm:"column:completed_at"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null;default:now()"`
}

func (recordingFinalizationJobRecord) TableName() string {
	return "recording_finalization_jobs"
}

type storageObjectRecord struct {
	ID               string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID        string          `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID          *string         `gorm:"column:robot_id;type:uuid;index"`
	RecordingChunkID *string         `gorm:"column:recording_chunk_id;type:uuid;index"`
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
	MissionID       *string         `gorm:"column:mission_id;type:uuid;index"`
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
	UserID         *string         `gorm:"column:user_id;type:uuid;index"`
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
	RobotID                *string         `gorm:"column:robot_id;type:uuid;index"`
	EventType              string          `gorm:"column:event_type;not null"`
	Severity               string          `gorm:"column:severity;not null"`
	Title                  string          `gorm:"column:title;not null"`
	Description            *string         `gorm:"column:description"`
	OccurredAt             time.Time       `gorm:"column:occurred_at;not null;index:events_mission_occurred_idx,sort:desc,priority:2"`
	Geom                   []byte          `gorm:"column:geom;type:geometry(Point,4326)"`
	RelatedStorageObjectID *string         `gorm:"column:related_storage_object_id;type:uuid;index"`
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
	RequestedBy   *string         `gorm:"column:requested_by;type:uuid;index"`
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
