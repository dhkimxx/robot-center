package port

type Store interface {
	RobotStore
	MissionStore
	SensorStore
	RecordingStore
	StreamSessionStore
	StorageAdminStore
}
