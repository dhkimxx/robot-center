package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
	"strings"
	"time"
)

func (s *Server) handleListSensorDescriptors(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	descriptors, err := s.services.Sensors.ListSensorDescriptors(r.Context(), missionID, robotCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sensorDescriptors": dto.SensorDescriptors(descriptors),
	})
}

func (s *Server) handleListSensorSamples(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	sensorID := strings.TrimSpace(r.URL.Query().Get("sensorId"))
	limit := intQueryValue(r, "limit", 100)
	samples, err := s.services.Sensors.ListSensorSamples(r.Context(), missionID, robotCode, sensorID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sensorSamples": dto.SensorSamples(samples),
	})
}

func (s *Server) handleListSensorLatest(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	latest, err := s.services.Sensors.ListLatestSensorSamples(r.Context(), missionID, robotCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"missionId": missionID,
		"robotCode": robotCode,
		"sensors":   dto.SensorLatest(latest),
	})
}

func (s *Server) handleCreateSensorSamples(w http.ResponseWriter, r *http.Request) {
	envelope, err := decodeSensorEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	samples, err := s.services.Sensors.SaveSensorEnvelope(r.Context(), envelope)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"sensorSamples": dto.SensorSamples(samples),
	})
}

type sensorDescriptorRequest struct {
	SensorID     string         `json:"sensorId"`
	ChannelRole  string         `json:"channelRole"`
	DisplayName  string         `json:"displayName"`
	Kind         string         `json:"kind"`
	SensorType   string         `json:"sensorType"`
	ValueType    string         `json:"valueType"`
	Unit         string         `json:"unit"`
	SamplingRate *float64       `json:"samplingRate"`
	SampleRateHz *float64       `json:"sampleRateHz"`
	Enabled      bool           `json:"enabled"`
	Metadata     map[string]any `json:"metadata"`
}

type sensorSampleRequest struct {
	SensorID     string         `json:"sensorId"`
	ChannelRole  string         `json:"channelRole"`
	MessageID    string         `json:"messageId"`
	Sequence     int64          `json:"sequence"`
	Timestamp    *time.Time     `json:"timestamp"`
	SentAt       *time.Time     `json:"sentAt"`
	NumericValue *float64       `json:"numericValue"`
	TextValue    string         `json:"textValue"`
	BoolValue    *bool          `json:"boolValue"`
	VectorValue  map[string]any `json:"vectorValue"`
	ObjectValue  map[string]any `json:"objectValue"`
	Values       any            `json:"values"`
	ObjectKey    string         `json:"objectKey"`
	RawPayload   map[string]any `json:"rawPayload"`
}

type sensorEnvelopeRequest struct {
	MessageID   string                    `json:"messageId"`
	MessageType string                    `json:"messageType"`
	RobotCode   string                    `json:"robotCode"`
	MissionID   string                    `json:"missionId"`
	ChannelRole string                    `json:"channelRole"`
	Sequence    int64                     `json:"sequence"`
	SentAt      *time.Time                `json:"sentAt"`
	Descriptors []sensorDescriptorRequest `json:"descriptors"`
	Samples     []sensorSampleRequest     `json:"samples"`
	Payload     map[string]any            `json:"payload"`
}

func decodeSensorEnvelope(r *http.Request) (domain.SensorEnvelope, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return domain.SensorEnvelope{}, err
	}
	var request sensorEnvelopeRequest
	if err := json.Unmarshal(rawPayload, &request); err != nil {
		return domain.SensorEnvelope{}, err
	}
	request.RobotCode = strings.TrimSpace(request.RobotCode)
	request.MissionID = strings.TrimSpace(request.MissionID)
	request.ChannelRole = strings.TrimSpace(request.ChannelRole)
	if request.RobotCode == "" {
		return domain.SensorEnvelope{}, errors.New("robotCode is required")
	}
	if request.MissionID == "" {
		return domain.SensorEnvelope{}, errors.New("missionId is required")
	}
	if request.ChannelRole == "" {
		request.ChannelRole = "channel.telemetry"
	}

	envelope := domain.SensorEnvelope{
		MessageID:   strings.TrimSpace(request.MessageID),
		MessageType: strings.TrimSpace(request.MessageType),
		RobotCode:   request.RobotCode,
		MissionID:   request.MissionID,
		ChannelRole: request.ChannelRole,
		Sequence:    request.Sequence,
		SentAt:      request.SentAt,
		ReceivedAt:  time.Now().UTC(),
		RawPayload:  append(json.RawMessage(nil), rawPayload...),
		Descriptors: make([]domain.SensorDescriptor, 0, len(request.Descriptors)),
		Samples:     make([]domain.SensorSample, 0, len(request.Samples)),
	}
	for _, descriptor := range request.Descriptors {
		sensorID := strings.TrimSpace(descriptor.SensorID)
		if sensorID == "" {
			continue
		}
		sampleRateHz := descriptor.SampleRateHz
		if sampleRateHz == nil {
			sampleRateHz = descriptor.SamplingRate
		}
		envelope.Descriptors = append(envelope.Descriptors, domain.SensorDescriptor{
			MissionID:    request.MissionID,
			RobotCode:    request.RobotCode,
			SensorID:     sensorID,
			ChannelRole:  utils.FirstNonEmptyString(descriptor.ChannelRole, request.ChannelRole),
			DisplayName:  utils.FirstNonEmptyString(descriptor.DisplayName, sensorID),
			SensorType:   inferSensorType(sensorID, utils.FirstNonEmptyString(descriptor.SensorType, descriptor.Kind)),
			ValueType:    utils.FirstNonEmptyString(descriptor.ValueType, "object"),
			Unit:         strings.TrimSpace(descriptor.Unit),
			SampleRateHz: sampleRateHz,
			Enabled:      descriptor.Enabled,
			Metadata:     utils.RawJSONOrEmpty(descriptor.Metadata),
		})
	}
	for _, sample := range request.Samples {
		sensorID := strings.TrimSpace(sample.SensorID)
		if sensorID == "" {
			continue
		}
		sentAt := sample.SentAt
		if sentAt == nil {
			sentAt = sample.Timestamp
		}
		if sentAt == nil {
			sentAt = request.SentAt
		}
		envelope.Samples = append(envelope.Samples, domain.SensorSample{
			MissionID:    request.MissionID,
			RobotCode:    request.RobotCode,
			SensorID:     sensorID,
			ChannelRole:  utils.FirstNonEmptyString(sample.ChannelRole, request.ChannelRole),
			MessageID:    utils.FirstNonEmptyString(sample.MessageID, request.MessageID),
			Sequence:     utils.FirstNonZeroInt64(sample.Sequence, request.Sequence),
			SentAt:       sentAt,
			ReceivedAt:   envelope.ReceivedAt,
			NumericValue: sample.NumericValue,
			TextValue:    strings.TrimSpace(sample.TextValue),
			BoolValue:    sample.BoolValue,
			VectorValue:  utils.RawJSONOrNil(sample.VectorValue),
			ObjectValue:  marshalSensorSampleObjectValue(sample),
			ObjectKey:    strings.TrimSpace(sample.ObjectKey),
			RawPayload:   utils.RawJSONOrEmpty(sample),
		})
	}
	if len(envelope.Descriptors) == 0 && len(envelope.Samples) == 0 && len(request.Payload) > 0 {
		envelope.Descriptors = append(envelope.Descriptors, domain.SensorDescriptor{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    "legacy.payload_1",
			ChannelRole: request.ChannelRole,
			DisplayName: "Legacy Payload",
			SensorType:  "legacy",
			ValueType:   "object",
			Enabled:     true,
			Metadata:    []byte("{}"),
		})
		envelope.Samples = append(envelope.Samples, domain.SensorSample{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    "legacy.payload_1",
			ChannelRole: request.ChannelRole,
			MessageID:   request.MessageID,
			Sequence:    request.Sequence,
			SentAt:      request.SentAt,
			ReceivedAt:  envelope.ReceivedAt,
			ObjectValue: utils.RawJSONOrEmpty(request.Payload),
			RawPayload:  envelope.RawPayload,
		})
	}
	return envelope, nil
}

func inferSensorType(sensorID string, explicitType string) string {
	if strings.TrimSpace(explicitType) != "" {
		return strings.TrimSpace(explicitType)
	}
	sensorID = strings.ToLower(strings.TrimSpace(sensorID))
	switch {
	case strings.Contains(sensorID, "position"):
		return "position"
	case strings.Contains(sensorID, "imu"):
		return "imu"
	case strings.Contains(sensorID, "odometry"):
		return "odometry"
	case strings.Contains(sensorID, "point_cloud"):
		return "point_cloud"
	case strings.Contains(sensorID, "battery"):
		return "battery"
	case strings.Contains(sensorID, "network"):
		return "network"
	case strings.Contains(sensorID, "temperature"):
		return "temperature"
	case strings.Contains(sensorID, "humidity"):
		return "humidity"
	case strings.Contains(sensorID, "gas"):
		return "gas"
	default:
		return "unknown"
	}
}

func marshalSensorSampleObjectValue(sample sensorSampleRequest) json.RawMessage {
	if sample.ObjectValue != nil {
		return utils.RawJSONOrNil(sample.ObjectValue)
	}
	if sample.Values != nil {
		return utils.RawJSONOrNil(sample.Values)
	}
	return nil
}
