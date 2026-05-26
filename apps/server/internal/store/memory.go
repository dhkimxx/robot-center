package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"robot-center/apps/server/internal/domain"
)

type MemoryStore struct {
	mu sync.RWMutex

	serverURL string

	robotSeq   int
	missionSeq int

	robotsByCode map[string]domain.Robot
	tokensByCode map[string]string
	tokenHashes  map[string]string

	missionsByCode           map[string]domain.Mission
	streamingByRobotCode     map[string]domain.StreamingStatus
	sensorDescriptors        map[string]domain.SensorDescriptor
	sensorSamplesByMissionID map[string][]domain.SensorSample
	recordingSessionsByKey   map[string]RecordingSession
	recordingChunksByID      map[string]domain.RecordingChunk
	recordingChunkKeyToID    map[string]string
}

func NewMemoryStore(serverURL string) *MemoryStore {
	return &MemoryStore{
		serverURL:                serverURL,
		robotsByCode:             map[string]domain.Robot{},
		tokensByCode:             map[string]string{},
		tokenHashes:              map[string]string{},
		missionsByCode:           map[string]domain.Mission{},
		streamingByRobotCode:     map[string]domain.StreamingStatus{},
		sensorDescriptors:        map[string]domain.SensorDescriptor{},
		sensorSamplesByMissionID: map[string][]domain.SensorSample{},
		recordingSessionsByKey:   map[string]RecordingSession{},
		recordingChunksByID:      map[string]domain.RecordingChunk{},
		recordingChunkKeyToID:    map[string]string{},
	}
}

func (s *MemoryStore) CreateRobot(_ context.Context, input CreateRobotInput) (domain.Robot, domain.RobotConnectionInfo, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, domain.RobotConnectionInfo{}, errors.New("displayName is required")
	}

	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.robotSeq++
	robotCode := fmt.Sprintf("robot-%03d", s.robotSeq)
	token := "rb_p0_" + randomHex(18)
	robot := domain.Robot{
		ID:          "rob_" + randomHex(12),
		RobotCode:   robotCode,
		DisplayName: input.DisplayName,
		ModelName:   input.ModelName,
		Status:      "offline",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.robotsByCode[robotCode] = robot
	s.tokensByCode[robotCode] = token
	s.tokenHashes[robotCode] = hashToken(token)

	return robot, domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  robotCode,
		RobotToken: token,
	}, nil
}

func (s *MemoryStore) ListRobots(_ context.Context) ([]domain.Robot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	robots := make([]domain.Robot, 0, len(s.robotsByCode))
	for _, robot := range s.robotsByCode {
		robots = append(robots, robot)
	}
	sort.Slice(robots, func(i, j int) bool {
		return robots[i].RobotCode < robots[j].RobotCode
	})
	return robots, nil
}

func (s *MemoryStore) UpdateRobot(_ context.Context, robotCode string, input UpdateRobotInput) (domain.Robot, error) {
	if input.DisplayName == "" {
		return domain.Robot{}, errors.New("displayName is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	robot, ok := s.robotsByCode[robotCode]
	if !ok {
		return domain.Robot{}, ErrNotFound
	}
	robot.DisplayName = input.DisplayName
	robot.ModelName = input.ModelName
	robot.UpdatedAt = time.Now().UTC()
	s.robotsByCode[robotCode] = robot
	return robot, nil
}

func (s *MemoryStore) ArchiveRobot(_ context.Context, robotCode string) (domain.Robot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	robot, ok := s.robotsByCode[robotCode]
	if !ok {
		return domain.Robot{}, ErrNotFound
	}
	for _, mission := range s.missionsByCode {
		if missionHasRobotCode(mission, robotCode) && (mission.Status == "ready" || mission.Status == "active") {
			return domain.Robot{}, ErrInvalidState
		}
	}
	robot.Status = "offline"
	robot.UpdatedAt = time.Now().UTC()
	delete(s.robotsByCode, robotCode)
	delete(s.tokensByCode, robotCode)
	delete(s.tokenHashes, robotCode)
	return robot, nil
}

func (s *MemoryStore) GetRobotConnectionInfo(_ context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.robotsByCode[robotCode]; !ok {
		return domain.RobotConnectionInfo{}, ErrNotFound
	}
	token := s.tokensByCode[robotCode]
	return domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  robotCode,
		RobotToken: token,
	}, nil
}

func (s *MemoryStore) RotateRobotConnectionToken(_ context.Context, robotCode string) (domain.RobotConnectionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.robotsByCode[robotCode]; !ok {
		return domain.RobotConnectionInfo{}, ErrNotFound
	}
	token := "rb_p0_" + randomHex(18)
	s.tokensByCode[robotCode] = token
	s.tokenHashes[robotCode] = hashToken(token)
	return domain.RobotConnectionInfo{
		ServerURL:  s.serverURL,
		RobotCode:  robotCode,
		RobotToken: token,
	}, nil
}

func (s *MemoryStore) CreateMission(_ context.Context, input CreateMissionInput) (domain.Mission, error) {
	if input.Name == "" {
		return domain.Mission{}, errors.New("name is required")
	}
	if input.MissionType == "" {
		return domain.Mission{}, errors.New("missionType is required")
	}

	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	robotCodes := normalizeMissionRobotCodes(input)
	for _, robotCode := range robotCodes {
		if _, ok := s.robotsByCode[robotCode]; !ok {
			return domain.Mission{}, fmt.Errorf("robotCode %s not found", robotCode)
		}
	}
	if conflicts := s.findActiveMissionConflictsLocked("", robotCodes); len(conflicts) > 0 {
		return domain.Mission{}, &MissionStartConflictError{Conflicts: conflicts}
	}

	s.missionSeq++
	missionCode := fmt.Sprintf("mission-%03d", s.missionSeq)
	mission := domain.Mission{
		ID:          "mis_" + randomHex(12),
		MissionCode: missionCode,
		Name:        input.Name,
		MissionType: input.MissionType,
		Status:      "ready",
		SiteNote:    input.SiteNote,
		RobotCode:   firstString(robotCodes),
		RobotCodes:  append([]string(nil), robotCodes...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.missionsByCode[missionCode] = mission
	return mission, nil
}

func (s *MemoryStore) ListMissions(_ context.Context) ([]domain.Mission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	missions := make([]domain.Mission, 0, len(s.missionsByCode))
	for _, mission := range s.missionsByCode {
		missions = append(missions, copyMission(mission))
	}
	sort.Slice(missions, func(i, j int) bool {
		return missions[i].CreatedAt.After(missions[j].CreatedAt)
	})
	return missions, nil
}

func (s *MemoryStore) StartMission(_ context.Context, missionCode string) (domain.Mission, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	mission, ok := s.missionsByCode[missionCode]
	if !ok {
		return domain.Mission{}, ErrNotFound
	}
	if mission.Status != "ready" {
		return domain.Mission{}, ErrInvalidState
	}
	if conflicts := s.findActiveMissionConflictsLocked(missionCode, missionRobotCodes(mission)); len(conflicts) > 0 {
		return domain.Mission{}, &MissionStartConflictError{Conflicts: conflicts}
	}
	mission.Status = "active"
	mission.StartedAt = &now
	mission.UpdatedAt = now
	s.missionsByCode[missionCode] = mission

	return copyMission(mission), nil
}

func (s *MemoryStore) findActiveMissionConflictsLocked(targetMissionCode string, robotCodes []string) []MissionStartConflict {
	if len(robotCodes) == 0 {
		return nil
	}
	targetRobots := map[string]struct{}{}
	for _, robotCode := range robotCodes {
		trimmed := strings.TrimSpace(robotCode)
		if trimmed != "" {
			targetRobots[trimmed] = struct{}{}
		}
	}
	if len(targetRobots) == 0 {
		return nil
	}

	targetMissionID := ""
	if targetMissionCode != "" {
		targetMissionID = s.missionsByCode[targetMissionCode].ID
	}
	conflicts := make([]MissionStartConflict, 0)
	seen := map[string]struct{}{}
	appendConflict := func(conflict MissionStartConflict) {
		key := conflict.RobotCode + "\x00" + conflict.ActiveMissionCode
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		conflicts = append(conflicts, conflict)
	}
	for _, mission := range s.missionsByCode {
		if mission.MissionCode == targetMissionCode || mission.Status != "active" {
			continue
		}
		for _, robotCode := range missionRobotCodes(mission) {
			if _, ok := targetRobots[robotCode]; !ok {
				continue
			}
			appendConflict(MissionStartConflict{
				RobotCode:         robotCode,
				ActiveMissionCode: mission.MissionCode,
			})
		}
	}
	now := time.Now().UTC()
	for robotCode := range targetRobots {
		status, ok := s.streamingByRobotCode[robotCode]
		if !ok || (status.Status != "streaming" && status.Status != "publishing") {
			continue
		}
		updatedAt := status.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = status.SentAt
		}
		if !updatedAt.IsZero() && (now.Sub(updatedAt) > streamingStatusFreshnessWindow || updatedAt.Sub(now) > streamingStatusFreshnessWindow) {
			continue
		}
		if targetMissionID != "" && status.MissionID == targetMissionID {
			continue
		}
		activeMissionCode := strings.TrimSpace(status.RoomID)
		for _, mission := range s.missionsByCode {
			if mission.ID == status.MissionID {
				activeMissionCode = mission.MissionCode
				break
			}
		}
		if activeMissionCode == "" {
			activeMissionCode = "streaming"
		}
		appendConflict(MissionStartConflict{
			RobotCode:         robotCode,
			ActiveMissionCode: activeMissionCode,
		})
	}
	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].RobotCode == conflicts[j].RobotCode {
			return conflicts[i].ActiveMissionCode < conflicts[j].ActiveMissionCode
		}
		return conflicts[i].RobotCode < conflicts[j].RobotCode
	})
	return conflicts
}

func (s *MemoryStore) EndMission(_ context.Context, missionCode string) (domain.Mission, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	mission, ok := s.missionsByCode[missionCode]
	if !ok {
		return domain.Mission{}, ErrNotFound
	}
	if mission.Status != "active" {
		return domain.Mission{}, ErrInvalidState
	}
	mission.Status = "ended"
	mission.EndedAt = &now
	mission.UpdatedAt = now
	s.missionsByCode[missionCode] = mission
	for _, robotCode := range missionRobotCodes(mission) {
		if status, ok := s.streamingByRobotCode[robotCode]; ok && status.MissionID == mission.ID {
			status.Status = "stopped"
			status.SentAt = now
			status.UpdatedAt = now
			s.streamingByRobotCode[robotCode] = status
		}
		robot := s.robotsByCode[robotCode]
		if robot.Status == "assigned" {
			robot.Status = "online"
			robot.UpdatedAt = now
			s.robotsByCode[robotCode] = robot
		}
	}
	return copyMission(mission), nil
}

func (s *MemoryStore) ApplyHeartbeat(_ context.Context, input HeartbeatInput, bearerToken string) (domain.Robot, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.authorizedLocked(input.RobotCode, bearerToken) {
		return domain.Robot{}, ErrUnauthorized
	}
	robot, ok := s.robotsByCode[input.RobotCode]
	if !ok {
		return domain.Robot{}, ErrNotFound
	}
	status := input.State
	if status == "" {
		status = "online"
	}
	robot.Status = normalizeRobotDeviceStatus(status)
	robot.LastSeenAt = &now
	robot.UpdatedAt = now
	s.robotsByCode[input.RobotCode] = robot
	return robot, nil
}

func (s *MemoryStore) FindActiveMissionForRobot(_ context.Context, robotCode string, bearerToken string) (domain.Mission, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.authorizedLocked(robotCode, bearerToken) {
		return domain.Mission{}, false, ErrUnauthorized
	}
	if _, ok := s.robotsByCode[robotCode]; !ok {
		return domain.Mission{}, false, ErrNotFound
	}

	for _, mission := range s.missionsByCode {
		if missionHasRobotCode(mission, robotCode) && mission.Status == "active" {
			return copyMission(mission), true, nil
		}
	}
	return domain.Mission{}, false, nil
}

func (s *MemoryStore) ApplyStreamingStatus(_ context.Context, status domain.StreamingStatus, bearerToken string) (domain.Robot, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	if status.SentAt.IsZero() {
		status.SentAt = now
	}
	status.RobotCode = strings.TrimSpace(status.RobotCode)
	status.MissionID = strings.TrimSpace(status.MissionID)
	status.RoomID = strings.TrimSpace(status.RoomID)
	status.Status = strings.TrimSpace(status.Status)
	if !s.authorizedLocked(status.RobotCode, bearerToken) {
		return domain.Robot{}, ErrUnauthorized
	}
	robot, ok := s.robotsByCode[status.RobotCode]
	if !ok {
		return domain.Robot{}, ErrNotFound
	}
	if isPublishingStatus(status.Status) {
		missionCode, ok := s.findActiveStreamingMissionCodeForRobotLocked(status.MissionID, status.RobotCode)
		if !ok || status.RoomID != missionCode {
			return domain.Robot{}, ErrInvalidState
		}
	}
	if isTerminalStreamingStatus(status.Status) {
		currentStatus, ok := s.streamingByRobotCode[status.RobotCode]
		if !ok || currentStatus.MissionID != status.MissionID || currentStatus.RoomID != status.RoomID {
			return robot, nil
		}
	}
	status.UpdatedAt = now
	robot.Status = normalizeRobotDeviceStatus(status.Status)
	robot.LastStreamingAt = &now
	robot.UpdatedAt = now
	s.robotsByCode[status.RobotCode] = robot
	s.streamingByRobotCode[status.RobotCode] = status
	return robot, nil
}

func (s *MemoryStore) ValidateActiveMissionRobot(_ context.Context, missionCode string, robotCode string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mission, ok := s.missionsByCode[strings.TrimSpace(missionCode)]
	if !ok || mission.Status != "active" || !missionHasRobotCode(mission, strings.TrimSpace(robotCode)) {
		return ErrInvalidState
	}
	return nil
}

func (s *MemoryStore) findActiveStreamingMissionCodeForRobotLocked(missionID string, robotCode string) (string, bool) {
	for _, mission := range s.missionsByCode {
		if mission.ID == missionID && mission.Status == "active" && missionHasRobotCode(mission, robotCode) {
			return mission.MissionCode, true
		}
	}
	return "", false
}

func (s *MemoryStore) ListStreamingStatuses(_ context.Context) ([]domain.StreamingStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]domain.StreamingStatus, 0, len(s.streamingByRobotCode))
	for _, status := range s.streamingByRobotCode {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].RobotCode < statuses[j].RobotCode
	})
	return statuses, nil
}

func (s *MemoryStore) SaveSensorEnvelope(_ context.Context, envelope domain.SensorEnvelope) ([]domain.SensorSample, error) {
	now := time.Now().UTC()
	envelope.RobotCode = strings.TrimSpace(envelope.RobotCode)
	envelope.MissionID = strings.TrimSpace(envelope.MissionID)
	envelope.ChannelRole = strings.TrimSpace(envelope.ChannelRole)
	if envelope.RobotCode == "" || envelope.MissionID == "" {
		return nil, ErrInvalidState
	}
	if envelope.ReceivedAt.IsZero() {
		envelope.ReceivedAt = now
	}
	if len(envelope.RawPayload) == 0 {
		envelope.RawPayload = []byte("{}")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, descriptor := range envelope.Descriptors {
		descriptor.SensorID = strings.TrimSpace(descriptor.SensorID)
		if descriptor.SensorID == "" {
			continue
		}
		key := sensorDescriptorKey(envelope.MissionID, envelope.RobotCode, descriptor.SensorID)
		existing := s.sensorDescriptors[key]
		if existing.ID == "" {
			existing.ID = "sdesc_" + randomHex(12)
			existing.FirstSeenAt = envelope.ReceivedAt
		}
		existing.MissionID = envelope.MissionID
		existing.RobotCode = envelope.RobotCode
		existing.SensorID = descriptor.SensorID
		existing.ChannelRole = firstNonEmpty(descriptor.ChannelRole, envelope.ChannelRole, "channel.telemetry")
		existing.DisplayName = firstNonEmpty(descriptor.DisplayName, descriptor.SensorID)
		existing.SensorType = firstNonEmpty(descriptor.SensorType, "unknown")
		existing.ValueType = firstNonEmpty(descriptor.ValueType, "object")
		existing.Unit = descriptor.Unit
		existing.SampleRateHz = descriptor.SampleRateHz
		existing.Enabled = descriptor.Enabled
		existing.Metadata = descriptor.Metadata
		if len(existing.Metadata) == 0 {
			existing.Metadata = []byte("{}")
		}
		existing.LastSeenAt = envelope.ReceivedAt
		s.sensorDescriptors[key] = existing
	}

	samples := make([]domain.SensorSample, 0, len(envelope.Samples))
	for _, sample := range envelope.Samples {
		sample.SensorID = strings.TrimSpace(sample.SensorID)
		if sample.SensorID == "" {
			continue
		}
		key := sensorDescriptorKey(envelope.MissionID, envelope.RobotCode, sample.SensorID)
		descriptor := s.sensorDescriptors[key]
		if descriptor.ID == "" {
			descriptor = createAutoSensorDescriptor(envelope, sample)
		} else {
			descriptor.LastSeenAt = envelope.ReceivedAt
		}
		s.sensorDescriptors[key] = descriptor
		if sample.ID == "" {
			sample.ID = "ssam_" + randomHex(12)
		}
		sample.DescriptorID = descriptor.ID
		sample.MissionID = envelope.MissionID
		sample.RobotCode = envelope.RobotCode
		sample.ChannelRole = firstNonEmpty(sample.ChannelRole, envelope.ChannelRole, "channel.telemetry")
		sample.MessageID = firstNonEmpty(sample.MessageID, envelope.MessageID)
		sample.Sequence = firstNonZeroInt64(sample.Sequence, envelope.Sequence)
		sample.SentAt = firstTimePointer(sample.SentAt, envelope.SentAt)
		if sample.ReceivedAt.IsZero() {
			sample.ReceivedAt = envelope.ReceivedAt
		}
		if len(sample.RawPayload) == 0 {
			sample.RawPayload = envelope.RawPayload
		}
		samples = append(samples, sample)
	}
	if len(samples) > 0 {
		items := append(s.sensorSamplesByMissionID[envelope.MissionID], samples...)
		s.sensorSamplesByMissionID[envelope.MissionID] = trimSensorSamples(items)
	}
	return samples, nil
}

func (s *MemoryStore) ListSensorDescriptors(_ context.Context, missionID string, robotCode string) ([]domain.SensorDescriptor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	descriptors := make([]domain.SensorDescriptor, 0)
	for _, descriptor := range s.sensorDescriptors {
		if descriptor.MissionID != missionID {
			continue
		}
		if strings.TrimSpace(robotCode) != "" && descriptor.RobotCode != strings.TrimSpace(robotCode) {
			continue
		}
		descriptors = append(descriptors, descriptor)
	}
	sort.Slice(descriptors, func(i, j int) bool {
		if descriptors[i].RobotCode == descriptors[j].RobotCode {
			return descriptors[i].SensorID < descriptors[j].SensorID
		}
		return descriptors[i].RobotCode < descriptors[j].RobotCode
	})
	return descriptors, nil
}

func (s *MemoryStore) ListSensorSamples(_ context.Context, missionID string, robotCode string, sensorID string, limit int) ([]domain.SensorSample, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]domain.SensorSample(nil), s.sensorSamplesByMissionID[missionID]...)
	filtered := make([]domain.SensorSample, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(robotCode) != "" && item.RobotCode != strings.TrimSpace(robotCode) {
			continue
		}
		if strings.TrimSpace(sensorID) != "" && item.SensorID != strings.TrimSpace(sensorID) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ReceivedAt.After(filtered[j].ReceivedAt)
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func (s *MemoryStore) ListLatestSensorSamples(_ context.Context, missionID string, robotCode string) ([]domain.SensorLatest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	latestByDescriptorKey := map[string]domain.SensorSample{}
	for _, sample := range s.sensorSamplesByMissionID[missionID] {
		if strings.TrimSpace(robotCode) != "" && sample.RobotCode != strings.TrimSpace(robotCode) {
			continue
		}
		key := sensorDescriptorKey(missionID, sample.RobotCode, sample.SensorID)
		current := latestByDescriptorKey[key]
		if current.ID == "" || sample.ReceivedAt.After(current.ReceivedAt) {
			latestByDescriptorKey[key] = sample
		}
	}
	latest := make([]domain.SensorLatest, 0)
	for _, descriptor := range s.sensorDescriptors {
		if descriptor.MissionID != missionID {
			continue
		}
		if strings.TrimSpace(robotCode) != "" && descriptor.RobotCode != strings.TrimSpace(robotCode) {
			continue
		}
		item := domain.SensorLatest{Descriptor: descriptor}
		if sample, ok := latestByDescriptorKey[sensorDescriptorKey(descriptor.MissionID, descriptor.RobotCode, descriptor.SensorID)]; ok {
			item.LatestSample = &sample
		}
		latest = append(latest, item)
	}
	sort.Slice(latest, func(i, j int) bool {
		if latest[i].Descriptor.RobotCode == latest[j].Descriptor.RobotCode {
			return latest[i].Descriptor.SensorID < latest[j].Descriptor.SensorID
		}
		return latest[i].Descriptor.RobotCode < latest[j].Descriptor.RobotCode
	})
	return latest, nil
}

func createAutoSensorDescriptor(envelope domain.SensorEnvelope, sample domain.SensorSample) domain.SensorDescriptor {
	return domain.SensorDescriptor{
		ID:          "sdesc_" + randomHex(12),
		MissionID:   envelope.MissionID,
		RobotCode:   envelope.RobotCode,
		SensorID:    sample.SensorID,
		ChannelRole: firstNonEmpty(sample.ChannelRole, envelope.ChannelRole, "channel.telemetry"),
		DisplayName: sample.SensorID,
		SensorType:  inferSensorTypeFromID(sample.SensorID),
		ValueType:   inferSensorValueType(sample),
		Enabled:     true,
		Metadata:    []byte(`{"source":"auto-sample"}`),
		FirstSeenAt: envelope.ReceivedAt,
		LastSeenAt:  envelope.ReceivedAt,
	}
}

func (s *MemoryStore) FindRecordingTarget(_ context.Context, missionCode string, robotCode string) (RecordingTarget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mission, ok := s.missionsByCode[missionCode]
	if !ok {
		return RecordingTarget{}, ErrNotFound
	}
	if robotCode == "" {
		robotCode = firstString(missionRobotCodes(mission))
	}
	if !missionHasRobotCode(mission, robotCode) {
		return RecordingTarget{}, ErrInvalidState
	}
	robot, ok := s.robotsByCode[robotCode]
	if !ok {
		return RecordingTarget{}, ErrNotFound
	}
	mission.RobotCode = robotCode
	return RecordingTarget{
		Mission:   copyMission(mission),
		RobotID:   robot.ID,
		RobotCode: robotCode,
	}, nil
}

func (s *MemoryStore) FindOrCreateRecordingSession(_ context.Context, missionID string, robotID string, _ int, startedAt time.Time) (RecordingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionKey := fmt.Sprintf("%s|%s", missionID, robotID)
	if session, ok := s.recordingSessionsByKey[sessionKey]; ok {
		return session, nil
	}
	session := RecordingSession{
		ID:        fmt.Sprintf("session-%s-%s", missionID, robotID),
		StartedAt: startedAt,
	}
	s.recordingSessionsByKey[sessionKey] = session
	return session, nil
}

func (s *MemoryStore) FindRecordingChunk(_ context.Context, recordingSessionID string, chunkIndex int) (domain.RecordingChunk, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, chunk := range s.recordingChunksByID {
		if chunk.RecordingSessionID == recordingSessionID && chunk.ChunkIndex == chunkIndex {
			return chunk, true, nil
		}
	}
	return domain.RecordingChunk{}, false, nil
}

func (s *MemoryStore) CreateRecordingChunk(_ context.Context, input CreateRecordingChunkInput) (domain.RecordingChunk, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chunkKey := fmt.Sprintf("%s|%s|%d", input.MissionCode, input.RobotCode, input.Window.Index)
	if chunkID := s.recordingChunkKeyToID[chunkKey]; chunkID != "" {
		return s.recordingChunksByID[chunkID], nil
	}
	chunk := domain.RecordingChunk{
		ID:                 "rec_" + randomHex(12),
		RecordingSessionID: input.RecordingSessionID,
		MissionID:          input.MissionID,
		MissionCode:        input.MissionCode,
		RobotCode:          input.RobotCode,
		ChunkIndex:         input.Window.Index,
		Status:             "recording",
		StartedAt:          input.Window.StartedAt,
		EndedAt:            input.Window.EndedAt,
		DurationSeconds:    input.Window.DurationSeconds,
		ManifestObjectKey:  input.MediaObjectKeys["manifest"],
		MediaObjectKeys:    input.MediaObjectKeys,
		AvailableFileTypes: map[string]bool{},
		CreatedAt:          input.CreatedAt,
		UpdatedAt:          input.UpdatedAt,
	}
	s.recordingChunkKeyToID[chunkKey] = chunk.ID
	s.recordingChunksByID[chunk.ID] = chunk
	return chunk, nil
}

func (s *MemoryStore) MarkRecordingChunkUploaded(_ context.Context, chunkID string, _ RecordingUploadMetadata) (domain.RecordingChunk, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	chunk, ok := s.recordingChunksByID[chunkID]
	if !ok {
		return domain.RecordingChunk{}, ErrNotFound
	}
	chunk.Status = "uploaded"
	if chunk.AvailableFileTypes == nil {
		chunk.AvailableFileTypes = map[string]bool{}
	}
	chunk.AvailableFileTypes["manifest"] = true
	chunk.UpdatedAt = now
	s.recordingChunksByID[chunk.ID] = chunk
	return chunk, nil
}

func (s *MemoryStore) MarkRecordingFileUploaded(_ context.Context, chunkID string, fileType string, _ RecordingUploadMetadata) (domain.RecordingChunk, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	chunk, ok := s.recordingChunksByID[chunkID]
	if !ok {
		return domain.RecordingChunk{}, ErrNotFound
	}
	if chunk.AvailableFileTypes == nil {
		chunk.AvailableFileTypes = map[string]bool{}
	}
	chunk.AvailableFileTypes[fileType] = true
	chunk.UpdatedAt = now
	s.recordingChunksByID[chunk.ID] = chunk
	return chunk, nil
}

func (s *MemoryStore) ListRecordingChunks(_ context.Context) ([]domain.RecordingChunk, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chunks := make([]domain.RecordingChunk, 0, len(s.recordingChunksByID))
	for _, chunk := range s.recordingChunksByID {
		chunks = append(chunks, chunk)
	}
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].StartedAt.Equal(chunks[j].StartedAt) {
			return chunks[i].MissionCode < chunks[j].MissionCode
		}
		return chunks[i].StartedAt.After(chunks[j].StartedAt)
	})
	return chunks, nil
}

func (s *MemoryStore) RecordingTargets(_ context.Context) ([]domain.Mission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targets := make([]domain.Mission, 0)
	for _, mission := range s.missionsByCode {
		if mission.Status != "active" {
			continue
		}
		for _, robotCode := range missionRobotCodes(mission) {
			target := copyMission(mission)
			target.RobotCode = robotCode
			targets = append(targets, target)
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].MissionCode < targets[j].MissionCode
	})
	return targets, nil
}

func normalizeMissionRobotCodes(input CreateMissionInput) []string {
	codes := make([]string, 0, len(input.RobotCodes)+1)
	if strings.TrimSpace(input.RobotCode) != "" {
		codes = append(codes, input.RobotCode)
	}
	codes = append(codes, input.RobotCodes...)
	return normalizeRobotCodes(codes)
}

func normalizeRobotCodes(codes []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(codes))
	for _, code := range codes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func copyMission(mission domain.Mission) domain.Mission {
	mission.RobotCodes = append([]string(nil), mission.RobotCodes...)
	return mission
}

func missionHasRobotCode(mission domain.Mission, robotCode string) bool {
	robotCode = strings.TrimSpace(robotCode)
	if robotCode == "" {
		return false
	}
	for _, assignedRobotCode := range missionRobotCodes(mission) {
		if assignedRobotCode == robotCode {
			return true
		}
	}
	return false
}

func missionRobotCodes(mission domain.Mission) []string {
	if len(mission.RobotCodes) > 0 {
		return mission.RobotCodes
	}
	if strings.TrimSpace(mission.RobotCode) == "" {
		return nil
	}
	return []string{strings.TrimSpace(mission.RobotCode)}
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func trimSensorSamples(items []domain.SensorSample) []domain.SensorSample {
	if len(items) <= 1000 {
		return items
	}
	return append([]domain.SensorSample(nil), items[len(items)-1000:]...)
}

func sensorDescriptorKey(missionID string, robotCode string, sensorID string) string {
	return strings.TrimSpace(missionID) + "|" + strings.TrimSpace(robotCode) + "|" + strings.TrimSpace(sensorID)
}

func (s *MemoryStore) authorizedLocked(robotCode string, bearerToken string) bool {
	if robotCode == "" || bearerToken == "" {
		return false
	}
	return s.tokenHashes[robotCode] == hashToken(bearerToken)
}
