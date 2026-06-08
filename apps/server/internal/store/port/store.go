package port

type Store interface {
	RobotStore
	MissionStore
	SensorStore
	EventStore
	RecordingStore
	StreamSessionStore
	StorageAdminStore
}
