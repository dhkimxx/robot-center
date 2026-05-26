package service

import (
	"sort"
	"strings"
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/sfu"
)

const defaultLiveStatusFreshnessWindow = 30 * time.Second

type LiveStatusService struct{}

type LiveStatusInput struct {
	Mission         domain.Mission
	Robots          []domain.Robot
	RecordingChunks []domain.RecordingChunk
	ObservedRooms   []sfu.ObservedRoomSummary
	Recorder        RecorderRuntimeSnapshot
	Now             time.Time
	FreshnessWindow time.Duration
}

type RecorderRuntimeSnapshot struct {
	Available bool
	Rooms     []RecorderRoomRuntime
}

type RecorderRoomRuntime struct {
	RoomID      string
	MissionCode string
	Robots      []RecorderRobotRuntime
}

type RecorderRobotRuntime struct {
	RobotCode        string
	TrackCount       int
	DataChannelCount int
	LastTrackAt      *time.Time
	LastDataAt       *time.Time
	LastPersistedAt  *time.Time
}

func (s *LiveStatusService) BuildMissionLiveStatus(input LiveStatusInput) domain.MissionLiveStatus {
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	freshnessWindow := input.FreshnessWindow
	if freshnessWindow <= 0 {
		freshnessWindow = defaultLiveStatusFreshnessWindow
	}

	robotsByCode := indexRobotsByCode(input.Robots)
	latestChunksByRobot := latestRecordingChunksByRobot(input.RecordingChunks, input.Mission.MissionCode)
	observedPublishersByRobot := observedPublishersForRoom(input.ObservedRooms, input.Mission.MissionCode)
	recorderRobotsByCode := recorderRuntimeForRoom(input.Recorder, input.Mission.MissionCode)

	robotCodes := missionRobotCodes(input.Mission)
	status := domain.MissionLiveStatus{
		MissionCode:   input.Mission.MissionCode,
		MissionStatus: input.Mission.Status,
		ObservedAt:    now,
		Robots:        make([]domain.RobotLiveStatus, 0, len(robotCodes)),
	}
	for _, robotCode := range robotCodes {
		robot := robotsByCode[robotCode]
		streamStatus := buildLiveStreamStatus(input.Mission, observedPublishersByRobot[robotCode], now, freshnessWindow)
		recordingStatus := buildLiveRecordingStatus(streamStatus, recorderRobotsByCode[robotCode], latestChunksByRobot[robotCode], input.Recorder.Available, now, freshnessWindow)
		status.Robots = append(status.Robots, domain.RobotLiveStatus{
			RobotCode:   robotCode,
			DisplayName: displayNameForRobot(robot, robotCode),
			Connection:  buildLiveConnectionStatus(robot, now, freshnessWindow),
			Stream:      streamStatus,
			Recording:   recordingStatus,
		})
	}
	return status
}

func indexRobotsByCode(robots []domain.Robot) map[string]domain.Robot {
	output := map[string]domain.Robot{}
	for _, robot := range robots {
		if strings.TrimSpace(robot.RobotCode) != "" {
			output[robot.RobotCode] = robot
		}
	}
	return output
}

func missionRobotCodes(mission domain.Mission) []string {
	seen := map[string]struct{}{}
	robotCodes := make([]string, 0, len(mission.RobotCodes)+1)
	addRobotCode := func(robotCode string) {
		robotCode = strings.TrimSpace(robotCode)
		if robotCode == "" {
			return
		}
		if _, ok := seen[robotCode]; ok {
			return
		}
		seen[robotCode] = struct{}{}
		robotCodes = append(robotCodes, robotCode)
	}
	for _, robotCode := range mission.RobotCodes {
		addRobotCode(robotCode)
	}
	addRobotCode(mission.RobotCode)
	sort.Strings(robotCodes)
	return robotCodes
}

func displayNameForRobot(robot domain.Robot, robotCode string) string {
	if strings.TrimSpace(robot.DisplayName) != "" {
		return robot.DisplayName
	}
	return robotCode
}

func buildLiveConnectionStatus(robot domain.Robot, now time.Time, freshnessWindow time.Duration) domain.LiveConnectionStatus {
	if robot.Status == "fault" {
		return domain.LiveConnectionStatus{State: "fault", Source: "robot_status", LastSeenAt: cloneDomainTimePointer(robot.LastSeenAt)}
	}
	if robot.Status == "offline" {
		return domain.LiveConnectionStatus{State: "offline", Source: "robot_status", LastSeenAt: cloneDomainTimePointer(robot.LastSeenAt)}
	}
	if robot.LastSeenAt != nil && isFreshTime(*robot.LastSeenAt, now, freshnessWindow) {
		return domain.LiveConnectionStatus{State: "online", Source: "heartbeat", LastSeenAt: cloneDomainTimePointer(robot.LastSeenAt)}
	}
	switch robot.Status {
	case "online", "streaming":
		return domain.LiveConnectionStatus{State: "online", Source: "robot_status", LastSeenAt: cloneDomainTimePointer(robot.LastSeenAt)}
	default:
		return domain.LiveConnectionStatus{State: "offline", Source: "unknown", LastSeenAt: cloneDomainTimePointer(robot.LastSeenAt)}
	}
}

func buildLiveStreamStatus(mission domain.Mission, publisher *sfu.ObservedPublisherSummary, now time.Time, freshnessWindow time.Duration) domain.LiveStreamStatus {
	status := domain.LiveStreamStatus{
		State:  "unknown",
		Source: "sfu",
		RoomID: mission.MissionCode,
	}
	if mission.Status != "active" {
		status.State = "ended"
		status.Reason = "mission_not_active"
		return status
	}
	if publisher == nil {
		status.State = "waiting"
		status.Reason = "no_publisher"
		return status
	}
	status.TrackCount = publisher.TrackCount
	status.DataChannelCount = publisher.DataChannelCount
	status.LastTrackAt = cloneDomainTimePointer(publisher.LastTrackAt)
	status.LastDataAt = cloneDomainTimePointer(publisher.LastDataAt)
	if isInactiveObservedICEState(publisher.ICEState) {
		status.State = "waiting"
		status.Reason = "publisher_" + publisher.ICEState
		return status
	}
	if isFreshPointer(publisher.LastTrackAt, now, freshnessWindow) || isFreshPointer(publisher.LastDataAt, now, freshnessWindow) {
		status.State = "streaming"
		return status
	}
	status.State = "waiting"
	status.Reason = "no_fresh_publisher"
	return status
}

func buildLiveRecordingStatus(streamStatus domain.LiveStreamStatus, recorderRobot *RecorderRobotRuntime, latestChunk *domain.RecordingChunk, recorderAvailable bool, now time.Time, freshnessWindow time.Duration) domain.LiveRecordingStatus {
	status := domain.LiveRecordingStatus{
		State:  "idle",
		Source: "recorder",
	}
	if latestChunk != nil {
		status.LatestChunk = recordingChunkSummary(*latestChunk)
		status.LatestChunkID = latestChunk.ID
		status.LatestChunkStatus = latestChunk.Status
	}
	if latestChunk != nil && latestChunk.Status == "failed" {
		status.State = "failed"
		status.Reason = "latest_chunk_failed"
		return status
	}
	if streamStatus.State != "streaming" {
		if latestChunk != nil && latestChunk.Status == "recording" && now.After(latestChunk.EndedAt) {
			status.State = "stale"
			status.Reason = "chunk_window_elapsed"
			return status
		}
		status.Reason = "no_active_stream"
		return status
	}
	if !recorderAvailable {
		status.Reason = "recorder_unavailable"
		return status
	}
	if recorderRobot == nil {
		status.Reason = "no_recorder_runtime"
		return status
	}
	if isFreshPointer(recorderRobot.LastTrackAt, now, freshnessWindow) || isFreshPointer(recorderRobot.LastDataAt, now, freshnessWindow) {
		status.State = "recording"
		return status
	}
	status.Reason = "recorder_runtime_stale"
	return status
}

func recordingChunkSummary(chunk domain.RecordingChunk) *domain.LiveRecordingChunkSummary {
	return &domain.LiveRecordingChunkSummary{
		ID:         chunk.ID,
		Status:     chunk.Status,
		StartedAt:  chunk.StartedAt,
		EndedAt:    chunk.EndedAt,
		UpdatedAt:  chunk.UpdatedAt,
		ChunkIndex: chunk.ChunkIndex,
	}
}

func latestRecordingChunksByRobot(chunks []domain.RecordingChunk, missionCode string) map[string]*domain.RecordingChunk {
	output := map[string]*domain.RecordingChunk{}
	for _, chunk := range chunks {
		if chunk.MissionCode != missionCode || strings.TrimSpace(chunk.RobotCode) == "" {
			continue
		}
		current := output[chunk.RobotCode]
		chunkCopy := chunk
		if current == nil || chunk.UpdatedAt.After(current.UpdatedAt) || chunk.StartedAt.After(current.StartedAt) {
			output[chunk.RobotCode] = &chunkCopy
		}
	}
	return output
}

func observedPublishersForRoom(rooms []sfu.ObservedRoomSummary, missionCode string) map[string]*sfu.ObservedPublisherSummary {
	output := map[string]*sfu.ObservedPublisherSummary{}
	for _, room := range rooms {
		if room.RoomID != missionCode {
			continue
		}
		for _, publisher := range room.Publishers {
			publisherCopy := publisher
			output[publisher.RobotCode] = &publisherCopy
		}
		return output
	}
	return output
}

func recorderRuntimeForRoom(snapshot RecorderRuntimeSnapshot, missionCode string) map[string]*RecorderRobotRuntime {
	output := map[string]*RecorderRobotRuntime{}
	for _, room := range snapshot.Rooms {
		if room.RoomID != missionCode && room.MissionCode != missionCode {
			continue
		}
		for _, robot := range room.Robots {
			robotCopy := robot
			output[robot.RobotCode] = &robotCopy
		}
		return output
	}
	return output
}

func isInactiveObservedICEState(state string) bool {
	switch strings.TrimSpace(state) {
	case "failed", "disconnected", "closed":
		return true
	default:
		return false
	}
}

func isFreshPointer(value *time.Time, now time.Time, freshnessWindow time.Duration) bool {
	return value != nil && isFreshTime(*value, now, freshnessWindow)
}

func isFreshTime(value time.Time, now time.Time, freshnessWindow time.Duration) bool {
	if value.IsZero() {
		return false
	}
	delta := now.Sub(value)
	if delta < 0 {
		delta = -delta
	}
	return delta <= freshnessWindow
}

func cloneDomainTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
