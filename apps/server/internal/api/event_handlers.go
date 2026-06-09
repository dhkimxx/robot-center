package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
	"robot-center/apps/server/internal/utils"
)

var canonicalDetectionTrackIDs = map[string]struct{}{
	"track.video_1": {},
	"track.video_2": {},
}

// @Summary 임무 이벤트 조회
// @Description operator가 mission 기준 이벤트 로그를 조회합니다. 기본 조회는 Live 이벤트 피드용이며 detection.object는 제외합니다. detection overlay 로그는 eventType 또는 includeDetections=true로 명시 조회합니다.
// @Tags Operator API
// @Produce json
// @Param missionCode path string true "임무 코드"
// @Param robotCode query string false "로봇 코드"
// @Param eventType query string false "이벤트 타입. 예: mission.event, detection.object"
// @Param eventCategory query string false "이벤트 카테고리. 예: mission, detection, alarm, system"
// @Param trackId query string false "미디어 track ID. 예: track.video_1"
// @Param includeDetections query bool false "기본 이벤트 피드에 detection.object 포함 여부"
// @Param limit query int false "조회 개수 제한. 기본 100, 최대 500"
// @Success 200 {object} dto.OperatorMissionEventsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/operator/missions/{missionCode}/events [get]
func (s *Server) handleListMissionEvents(w http.ResponseWriter, r *http.Request) {
	missionCode := strings.TrimSpace(r.PathValue("missionCode"))
	query := store.EventQuery{
		MissionID:         missionCode,
		RobotCode:         strings.TrimSpace(r.URL.Query().Get("robotCode")),
		EventType:         strings.TrimSpace(r.URL.Query().Get("eventType")),
		EventCategory:     strings.TrimSpace(r.URL.Query().Get("eventCategory")),
		TrackID:           strings.TrimSpace(r.URL.Query().Get("trackId")),
		IncludeDetections: boolQueryValue(r, "includeDetections", false),
		Limit:             intQueryValue(r, "limit", 100),
	}
	events, err := s.services.Events.ListMissionEvents(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.OperatorMissionEventsPayload(missionCode, events))
}

// @Summary recorder 이벤트 저장
// @Description recorder-worker가 channel.event envelope를 저장합니다. Live overlay는 별도 projection이고, 이 API는 append-only 이벤트 로그를 남깁니다. detection.object는 values.trackId가 track.video_1 또는 track.video_2여야 하며, values.detections[].className/confidence/bbox.x/y/width/height를 검증합니다. mission.event는 values.title이 필수이고 values.severity는 info, warning, critical만 사용합니다.
// @Tags Recorder API
// @Accept json
// @Produce json
// @Param request body dto.EventEnvelopeRequest true "channel.event envelope"
// @Success 201 {object} dto.MissionEventsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recorder/events [post]
func (s *Server) handleCreateMissionEvents(w http.ResponseWriter, r *http.Request) {
	envelope, err := decodeMissionEventEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	events, err := s.services.Events.SaveMissionEventEnvelope(r.Context(), envelope)
	if err != nil {
		if errors.Is(err, store.ErrInvalidState) || errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.MissionEventsPayload(events))
}

func decodeMissionEventEnvelope(r *http.Request) (domain.MissionEventEnvelope, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return domain.MissionEventEnvelope{}, err
	}
	var request dto.EventEnvelopeRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return domain.MissionEventEnvelope{}, err
	}

	robotCode := strings.TrimSpace(request.RobotCode)
	missionID := utils.FirstNonEmptyString(request.MissionID, request.MissionCode)
	channelRole := utils.FirstNonEmptyString(request.ChannelRole, domain.EventSourceChannel)
	if robotCode == "" {
		return domain.MissionEventEnvelope{}, errors.New("robotCode is required")
	}
	if missionID == "" {
		return domain.MissionEventEnvelope{}, errors.New("missionId is required")
	}
	if channelRole != domain.EventSourceChannel {
		return domain.MissionEventEnvelope{}, errors.New("channelRole must be channel.event")
	}
	if len(request.Events) == 0 {
		return domain.MissionEventEnvelope{}, errors.New("events are required")
	}

	receivedAt := time.Now().UTC()
	envelope := domain.MissionEventEnvelope{
		MessageID:   strings.TrimSpace(request.MessageID),
		MessageType: strings.TrimSpace(request.MessageType),
		RobotCode:   robotCode,
		MissionID:   missionID,
		ReceivedAt:  receivedAt,
		Events:      make([]domain.MissionEvent, 0, len(request.Events)),
		RawMessage:  append(json.RawMessage(nil), rawPayload...),
	}
	for _, item := range request.Events {
		event, err := missionEventFromRequest(item, envelope)
		if err != nil {
			return domain.MissionEventEnvelope{}, err
		}
		envelope.Events = append(envelope.Events, event)
	}
	return envelope, nil
}

func missionEventFromRequest(item dto.EventItemRequest, envelope domain.MissionEventEnvelope) (domain.MissionEvent, error) {
	eventType := strings.TrimSpace(item.EventType)
	if eventType == "" {
		return domain.MissionEvent{}, errors.New("eventType is required")
	}
	values, err := normalizeEventValues(item.Values)
	if err != nil {
		return domain.MissionEvent{}, err
	}
	timestamp := envelope.ReceivedAt
	if item.Timestamp != nil {
		timestamp = item.Timestamp.UTC()
	}
	trackID, detectionCount, err := eventValueProjection(eventType, values)
	if err != nil {
		return domain.MissionEvent{}, err
	}
	commonValues, err := parseEventCommonValues(eventType, values)
	if err != nil {
		return domain.MissionEvent{}, err
	}
	title := utils.FirstNonEmptyString(commonValues.Title, eventType)
	return domain.MissionEvent{
		MissionID:      envelope.MissionID,
		RobotCode:      envelope.RobotCode,
		EventID:        strings.TrimSpace(item.EventID),
		EventType:      eventType,
		EventCategory:  eventCategoryForType(eventType),
		TrackID:        trackID,
		Severity:       commonValues.Severity,
		Title:          title,
		Description:    commonValues.Description,
		Timestamp:      timestamp,
		ReceivedAt:     envelope.ReceivedAt,
		DetectionCount: detectionCount,
		Values:         values,
		RawMessage:     envelope.RawMessage,
	}, nil
}

func eventValueProjection(eventType string, values json.RawMessage) (string, *int, error) {
	if eventType != domain.EventTypeDetectionObject {
		return "", nil, nil
	}
	detectionValues, err := parseDetectionObjectValues(values)
	if err != nil {
		return "", nil, err
	}
	detectionCount := len(detectionValues.Detections)
	return detectionValues.TrackID, &detectionCount, nil
}

type detectionObjectValues struct {
	TrackID    string                     `json:"trackId"`
	Detections []detectionObjectDetection `json:"detections"`
}

type detectionObjectDetection struct {
	ClassName  string               `json:"className"`
	Confidence *float64             `json:"confidence"`
	BBox       *detectionObjectBbox `json:"bbox"`
}

type detectionObjectBbox struct {
	X      *float64 `json:"x"`
	Y      *float64 `json:"y"`
	Width  *float64 `json:"width"`
	Height *float64 `json:"height"`
}

func parseDetectionObjectValues(values json.RawMessage) (detectionObjectValues, error) {
	var request detectionObjectValues
	if err := json.Unmarshal(values, &request); err != nil {
		return detectionObjectValues{}, err
	}
	request.TrackID = strings.TrimSpace(request.TrackID)
	if request.TrackID == "" {
		return detectionObjectValues{}, errors.New("values.trackId is required for detection.object")
	}
	if _, ok := canonicalDetectionTrackIDs[request.TrackID]; !ok {
		return detectionObjectValues{}, errors.New("values.trackId must be track.video_1 or track.video_2 for detection.object")
	}
	if request.Detections == nil {
		return detectionObjectValues{}, errors.New("values.detections is required for detection.object")
	}
	for index, detection := range request.Detections {
		if err := validateDetectionObjectDetection(index, detection); err != nil {
			return detectionObjectValues{}, err
		}
	}
	return request, nil
}

func validateDetectionObjectDetection(index int, detection detectionObjectDetection) error {
	if strings.TrimSpace(detection.ClassName) == "" {
		return fmt.Errorf("values.detections[%d].className is required for detection.object", index)
	}
	if detection.Confidence == nil {
		return fmt.Errorf("values.detections[%d].confidence is required for detection.object", index)
	}
	if *detection.Confidence < 0 || *detection.Confidence > 1 {
		return fmt.Errorf("values.detections[%d].confidence must be between 0 and 1 for detection.object", index)
	}
	if err := validateDetectionObjectBbox(index, detection.BBox); err != nil {
		return err
	}
	return nil
}

func validateDetectionObjectBbox(index int, bbox *detectionObjectBbox) error {
	if bbox == nil {
		return fmt.Errorf("values.detections[%d].bbox is required for detection.object", index)
	}
	x, err := requiredUnitFloat(bbox.X, fmt.Sprintf("values.detections[%d].bbox.x", index))
	if err != nil {
		return err
	}
	y, err := requiredUnitFloat(bbox.Y, fmt.Sprintf("values.detections[%d].bbox.y", index))
	if err != nil {
		return err
	}
	width, err := requiredUnitFloat(bbox.Width, fmt.Sprintf("values.detections[%d].bbox.width", index))
	if err != nil {
		return err
	}
	height, err := requiredUnitFloat(bbox.Height, fmt.Sprintf("values.detections[%d].bbox.height", index))
	if err != nil {
		return err
	}
	if width <= 0 {
		return fmt.Errorf("values.detections[%d].bbox.width must be greater than 0 for detection.object", index)
	}
	if height <= 0 {
		return fmt.Errorf("values.detections[%d].bbox.height must be greater than 0 for detection.object", index)
	}
	if x+width > 1 {
		return fmt.Errorf("values.detections[%d].bbox.x + width must be <= 1 for detection.object", index)
	}
	if y+height > 1 {
		return fmt.Errorf("values.detections[%d].bbox.y + height must be <= 1 for detection.object", index)
	}
	return nil
}

func requiredUnitFloat(value *float64, fieldName string) (float64, error) {
	if value == nil {
		return 0, fmt.Errorf("%s is required for detection.object", fieldName)
	}
	if *value < 0 || *value > 1 {
		return 0, fmt.Errorf("%s must be between 0 and 1 for detection.object", fieldName)
	}
	return *value, nil
}

func normalizeEventValues(values json.RawMessage) (json.RawMessage, error) {
	if len(values) == 0 || string(values) == "null" {
		return nil, errors.New("values is required")
	}
	if !json.Valid(values) {
		return nil, errors.New("values must be a valid JSON object")
	}
	var object map[string]any
	if err := json.Unmarshal(values, &object); err != nil {
		return nil, errors.New("values must be a JSON object")
	}
	return append(json.RawMessage(nil), values...), nil
}

type eventCommonValues struct {
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func parseEventCommonValues(eventType string, values json.RawMessage) (eventCommonValues, error) {
	if eventType == domain.EventTypeMissionEvent {
		return parseMissionEventValues(values)
	}
	var request eventCommonValues
	if err := json.Unmarshal(values, &request); err != nil {
		return eventCommonValues{Severity: "info"}, nil
	}
	request.Severity = normalizeEventSeverity(request.Severity)
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	return request, nil
}

func parseMissionEventValues(values json.RawMessage) (eventCommonValues, error) {
	var request eventCommonValues
	decoder := json.NewDecoder(bytes.NewReader(values))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return eventCommonValues{}, errors.New("mission.event values must only contain severity, title, and description")
	}
	request.Severity = normalizeEventSeverity(request.Severity)
	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" {
		return eventCommonValues{}, errors.New("values.title is required for mission.event")
	}
	request.Description = strings.TrimSpace(request.Description)
	return request, nil
}

func normalizeEventSeverity(severity string) string {
	normalized := strings.ToLower(strings.TrimSpace(severity))
	switch normalized {
	case "warning", "critical":
		return normalized
	default:
		return "info"
	}
}

func eventCategoryForType(eventType string) string {
	switch {
	case eventType == domain.EventTypeDetectionObject:
		return domain.EventCategoryDetection
	case eventType == domain.EventTypeMissionEvent:
		return domain.EventCategoryMission
	case strings.HasPrefix(eventType, "alarm."):
		return domain.EventCategoryAlarm
	case strings.HasPrefix(eventType, "system."):
		return domain.EventCategorySystem
	default:
		return domain.EventCategoryMission
	}
}
