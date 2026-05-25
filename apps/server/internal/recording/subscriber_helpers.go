package recording

import (
	"encoding/json"
	"net/url"
	"sort"
	"strings"

	"github.com/pion/webrtc/v4"

	"robot-center/apps/server/internal/domain"
)

func buildSignalingURL(baseURL string, roomID string, role string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	query := parsed.Query()
	query.Set("room", roomID)
	query.Set("role", role)
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
		return "telemetry"
	case "channel.event", "channel.spatial", "channel.control":
		return ""
	default:
		return strings.TrimSpace(label)
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

func recorderDataChannelPayloadWithRobotCode(robotCode string, payload []byte) []byte {
	if strings.TrimSpace(robotCode) == "" || !json.Valid(payload) {
		return payload
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return payload
	}
	if _, ok := object["robotCode"]; ok {
		return payload
	}
	object["robotCode"] = strings.TrimSpace(robotCode)
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
	if strings.Contains(raw, "track.video_1") {
		return "track.video_1"
	}
	if strings.Contains(raw, "track.video_2") {
		return "track.video_2"
	}
	if strings.Contains(raw, "track.audio_1") {
		return "track.audio_1"
	}
	if strings.Contains(raw, "track.audio_2") {
		return "track.audio_2"
	}
	if strings.Contains(raw, "thermal") {
		return "thermal"
	}
	if strings.Contains(raw, "rgb") {
		return "rgb"
	}
	if track.Kind() == webrtc.RTPCodecTypeAudio {
		return "audio"
	}
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		return "video"
	}
	return track.Kind().String()
}
