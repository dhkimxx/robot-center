package recording

import (
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
)

const recorderUnmappedTrackPrefix = "unmapped."

func buildSignalingURL(baseURL string, roomID string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	query := parsed.Query()
	query.Set("room", roomID)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func groupRecordingTargetsByMission(targets []domain.Mission) map[string][]domain.Mission {
	grouped := map[string][]domain.Mission{}
	for _, target := range targets {
		roomID := recorderSignalingRoomID(target.MissionCode)
		robotCode := strings.TrimSpace(target.RobotCode)
		if roomID == "" || robotCode == "" {
			continue
		}
		target.MissionCode = roomID
		target.RobotCode = robotCode
		grouped[roomID] = append(grouped[roomID], target)
	}
	return grouped
}

func recorderSignalingRoomID(missionCode string) string {
	return strings.TrimSpace(missionCode)
}

func recorderMediaKey(missionCode string, robotCode string) string {
	return domain.StreamRoomID(strings.TrimSpace(missionCode), strings.TrimSpace(robotCode))
}

func splitRecorderMediaKey(mediaKey string) (string, string) {
	parts := strings.SplitN(mediaKey, "__", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func robotCodesFromTargets(targets []domain.Mission) []string {
	seen := map[string]struct{}{}
	robotCodes := make([]string, 0, len(targets))
	for _, target := range targets {
		robotCode := strings.TrimSpace(target.RobotCode)
		if robotCode == "" {
			continue
		}
		if _, ok := seen[robotCode]; ok {
			continue
		}
		seen[robotCode] = struct{}{}
		robotCodes = append(robotCodes, robotCode)
	}
	sort.Strings(robotCodes)
	return robotCodes
}

func robotCodeSet(robotCodes []string) map[string]struct{} {
	output := map[string]struct{}{}
	for _, robotCode := range robotCodes {
		if strings.TrimSpace(robotCode) != "" {
			output[strings.TrimSpace(robotCode)] = struct{}{}
		}
	}
	return output
}

func firstRobotCode(robotCodes []string) string {
	if len(robotCodes) == 0 {
		return ""
	}
	return robotCodes[0]
}

func sortedRobotCodes(robotCodes map[string]struct{}) []string {
	output := make([]string, 0, len(robotCodes))
	for robotCode := range robotCodes {
		output = append(output, robotCode)
	}
	sort.Strings(output)
	return output
}

func ensureRecorderRobotRuntime(status *recorderSessionStatus, robotCode string) recorderRobotRuntime {
	if status.robotStatuses == nil {
		status.robotStatuses = map[string]recorderRobotRuntime{}
	}
	runtime := status.robotStatuses[robotCode]
	if runtime.trackLabels == nil {
		runtime.trackLabels = map[string]struct{}{}
	}
	if runtime.dataChannelLabels == nil {
		runtime.dataChannelLabels = map[string]struct{}{}
	}
	return runtime
}

func ensureRecorderDataChannelRuntime(status *recorderSessionStatus, label string, observedAt time.Time) recorderDataChannelRuntime {
	if status.dataChannelStates == nil {
		status.dataChannelStates = map[string]recorderDataChannelRuntime{}
	}
	if status.dataChannelLabels == nil {
		status.dataChannelLabels = map[string]struct{}{}
	}
	status.dataChannelLabels[label] = struct{}{}
	runtime := status.dataChannelStates[label]
	if runtime.label == "" {
		runtime.label = label
		runtime.state = "detected"
		runtime.detectedAt = observedAt
	}
	if runtime.state == "" {
		runtime.state = "detected"
	}
	return runtime
}

func subscriberDataChannelStatuses(status recorderSessionStatus) []SubscriberDataChannelStatus {
	labels := make([]string, 0, len(status.dataChannelStates))
	for label := range status.dataChannelStates {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	output := make([]SubscriberDataChannelStatus, 0, len(labels))
	for _, label := range labels {
		runtime := status.dataChannelStates[label]
		output = append(output, SubscriberDataChannelStatus{
			Label:         label,
			State:         runtime.state,
			DetectedAt:    recorderTimePointer(runtime.detectedAt),
			OpenedAt:      recorderTimePointer(runtime.openedAt),
			LastMessageAt: recorderTimePointer(runtime.lastMessageAt),
			MessageCount:  runtime.messageCount,
			ClosedAt:      recorderTimePointer(runtime.closedAt),
			LastError:     runtime.lastError,
		})
	}
	return output
}

func recorderTimePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func (w *Worker) markRecorderDataChannelDetected(roomID string, label string) {
	observedAt := time.Now().UTC()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		runtime := ensureRecorderDataChannelRuntime(status, label, observedAt)
		status.dataChannelStates[label] = runtime
		status.lastDataLabel = label
	})
}

func (w *Worker) markRecorderDataChannelOpen(roomID string, label string) {
	observedAt := time.Now().UTC()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		runtime := ensureRecorderDataChannelRuntime(status, label, observedAt)
		runtime.state = "open"
		runtime.openedAt = observedAt
		runtime.lastError = ""
		status.dataChannelStates[label] = runtime
		status.lastDataLabel = label
	})
}

func (w *Worker) markRecorderDataChannelMessage(roomID string, label string, observedAt time.Time) {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		runtime := ensureRecorderDataChannelRuntime(status, label, observedAt)
		if runtime.state != "closed" && runtime.state != "error" {
			runtime.state = "open"
		}
		runtime.lastMessageAt = observedAt
		runtime.messageCount++
		status.dataChannelStates[label] = runtime
		status.dataMessageCount++
		status.lastDataLabel = label
		status.lastDataMessageAt = observedAt
	})
}

func (w *Worker) markRecorderDataChannelClosed(roomID string, label string) {
	observedAt := time.Now().UTC()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		runtime := ensureRecorderDataChannelRuntime(status, label, observedAt)
		runtime.state = "closed"
		runtime.closedAt = observedAt
		status.dataChannelStates[label] = runtime
	})
}

func (w *Worker) markRecorderDataChannelError(roomID string, label string, err error) {
	observedAt := time.Now().UTC()
	w.updateSubscriberStatus(roomID, func(status *recorderSessionStatus) {
		runtime := ensureRecorderDataChannelRuntime(status, label, observedAt)
		if runtime.state != "closed" {
			runtime.state = "error"
		}
		if err != nil {
			runtime.lastError = err.Error()
		}
		status.dataChannelStates[label] = runtime
	})
}

func (w *Worker) markRecorderRobotTrackActivity(mediaKey string, label string, observedAt time.Time) {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	missionCode, robotCode := splitRecorderMediaKey(mediaKey)
	if missionCode == "" || robotCode == "" {
		return
	}
	w.updateSubscriberStatus(missionCode, func(status *recorderSessionStatus) {
		status.robotCodes[robotCode] = struct{}{}
		if status.robotCode == "" {
			status.robotCode = robotCode
		}
		robotStatus := ensureRecorderRobotRuntime(status, robotCode)
		trackLabel := strings.TrimSpace(label)
		if trackLabel != "" {
			robotStatus.trackLabels[trackLabel] = struct{}{}
			status.trackLabels[recorderTrackLabel(robotCode, trackLabel)] = struct{}{}
			status.lastTrackLabel = recorderTrackLabel(robotCode, trackLabel)
		}
		robotStatus.lastTrackAt = observedAt
		robotStatus.updatedAt = observedAt
		status.robotStatuses[robotCode] = robotStatus
	})
}

func subscriberRobotStatuses(status recorderSessionStatus) []SubscriberRobotStatus {
	robotCodes := sortedRobotCodes(status.robotCodes)
	output := make([]SubscriberRobotStatus, 0, len(robotCodes))
	for _, robotCode := range robotCodes {
		runtime := status.robotStatuses[robotCode]
		output = append(output, SubscriberRobotStatus{
			RobotCode:        robotCode,
			TrackCount:       recorderCanonicalTrackCount(runtime.trackLabels),
			DataChannelCount: len(runtime.dataChannelLabels),
			LastTrackAt:      runtime.lastTrackAt,
			LastDataAt:       runtime.lastDataAt,
			LastPersistedAt:  runtime.lastPersistedAt,
			UpdatedAt:        runtime.updatedAt,
		})
	}
	return output
}

func subscriberRoomTrackCount(status recorderSessionStatus) int {
	if len(status.robotStatuses) == 0 {
		return recorderCanonicalTrackCount(status.trackLabels)
	}
	count := 0
	for _, runtime := range status.robotStatuses {
		count += recorderCanonicalTrackCount(runtime.trackLabels)
	}
	return count
}

func recorderCanonicalTrackCount(labels map[string]struct{}) int {
	count := 0
	for label := range labels {
		if isRecorderCanonicalTrackLabel(label) {
			count++
		}
	}
	return count
}

func isRecorderCanonicalTrackLabel(label string) bool {
	normalized := strings.TrimSpace(label)
	if index := strings.Index(normalized, ":"); index >= 0 {
		normalized = strings.TrimSpace(normalized[index+1:])
	}
	switch normalized {
	case "track.video_1", "track.video_2", "track.audio_1", "track.audio_2":
		return true
	default:
		return false
	}
}

func (w *Worker) singleSubscriberRobotCode(roomID string) string {
	w.subscriberMu.RLock()
	defer w.subscriberMu.RUnlock()

	statusKey := roomID
	if _, ok := w.subscriberStatuses[statusKey]; !ok {
		if missionCode, _ := splitRecorderMediaKey(roomID); missionCode != "" {
			statusKey = missionCode
		}
	}
	status := w.subscriberStatuses[statusKey]
	if len(status.robotCodes) == 1 {
		for robotCode := range status.robotCodes {
			return robotCode
		}
	}
	return ""
}

func (w *Worker) closeRecorderSessionAudioWriters(roomID string) {
	w.subscriberMu.RLock()
	status := w.subscriberStatuses[roomID]
	robotCodes := sortedRobotCodes(status.robotCodes)
	w.subscriberMu.RUnlock()

	for _, robotCode := range robotCodes {
		w.closeAudioWriter(recorderMediaKey(roomID, robotCode))
	}
}

func classifyRecorderTrack(track *webrtc.TrackRemote) (string, string) {
	label := classifyTrack(track)
	trackID := strings.TrimSpace(track.ID())
	if label != "" && strings.HasSuffix(trackID, "-"+label) {
		return strings.TrimSuffix(trackID, "-"+label), label
	}
	streamID := strings.TrimSpace(track.StreamID())
	if strings.HasPrefix(streamID, "robot-") {
		return strings.TrimPrefix(streamID, "robot-"), label
	}
	return "", label
}

func recorderStorageDataChannelLabel(label string) string {
	switch strings.TrimSpace(label) {
	case "channel.telemetry":
		return "channel.telemetry"
	case "channel.event":
		return "channel.event"
	case "channel.control":
		return ""
	default:
		return ""
	}
}

func recorderDataChannelFileLabel(storageLabel string) string {
	switch strings.TrimSpace(storageLabel) {
	case "channel.telemetry":
		return "telemetry"
	default:
		return ""
	}
}

func recorderTrackLabel(robotCode string, label string) string {
	if strings.TrimSpace(robotCode) == "" {
		return strings.TrimSpace(label)
	}
	return strings.TrimSpace(robotCode) + ":" + strings.TrimSpace(label)
}

func robotCodeFromDataPayload(payload []byte) string {
	if !json.Valid(payload) {
		return ""
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		return ""
	}
	robotCode, _ := object["robotCode"].(string)
	return strings.TrimSpace(robotCode)
}

func recorderDataChannelPayloadWithContext(roomID string, robotCode string, channelRole string, payload []byte) []byte {
	if !json.Valid(payload) {
		return payload
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return payload
	}
	if strings.TrimSpace(robotCode) != "" {
		object["robotCode"] = strings.TrimSpace(robotCode)
	}
	if strings.TrimSpace(roomID) != "" {
		object["missionId"] = strings.TrimSpace(roomID)
		object["missionCode"] = strings.TrimSpace(roomID)
	}
	if strings.TrimSpace(channelRole) != "" {
		object["channelRole"] = strings.TrimSpace(channelRole)
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		return payload
	}
	return encoded
}

func shouldIgnoreTargetedMessage(payload map[string]any, selfPeerID string) bool {
	targetPeerID := payloadString(payload, "targetPeerId")
	return targetPeerID != "" && selfPeerID != "" && targetPeerID != selfPeerID
}

func payloadString(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func payloadStringPointer(payload map[string]any, key string) *string {
	value := payloadString(payload, key)
	if value == "" {
		return nil
	}
	return &value
}

func payloadUint16Pointer(payload map[string]any, key string) *uint16 {
	value, ok := payload[key]
	if !ok {
		return nil
	}
	var number uint16
	switch typed := value.(type) {
	case float64:
		number = uint16(typed)
	case int:
		number = uint16(typed)
	default:
		return nil
	}
	return &number
}

func dereferenceString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func dereferenceUint16(value *uint16) uint16 {
	if value == nil {
		return 0
	}
	return *value
}

func classifyTrack(track *webrtc.TrackRemote) string {
	raw := strings.ToLower(track.StreamID() + " " + track.ID() + " " + track.Codec().MimeType)
	for _, label := range []string{"track.video_1", "track.video_2", "track.audio_1", "track.audio_2"} {
		if strings.Contains(raw, label) {
			return label
		}
	}
	return recorderUnmappedTrackLabel(track)
}

func recorderUnmappedTrackLabel(track *webrtc.TrackRemote) string {
	for _, candidate := range []string{track.ID(), track.StreamID()} {
		if index := strings.Index(candidate, recorderUnmappedTrackPrefix); index >= 0 {
			return utils.SafeTrackToken(candidate[index:])
		}
	}
	for _, candidate := range []string{track.ID(), track.StreamID(), track.Kind().String()} {
		token := utils.SafeTrackToken(candidate)
		if token != "unknown" {
			return recorderUnmappedTrackPrefix + token
		}
	}
	return recorderUnmappedTrackPrefix + "unknown"
}
