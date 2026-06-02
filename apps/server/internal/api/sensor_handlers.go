package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"robot-center/apps/server/internal/api/dto"
	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
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
	writeJSON(w, http.StatusOK, dto.SensorDescriptorsPayload(descriptors))
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
	writeJSON(w, http.StatusOK, dto.SensorSamplesPayload(samples))
}

func (s *Server) handleListSensorLatest(w http.ResponseWriter, r *http.Request) {
	missionID := strings.TrimSpace(r.URL.Query().Get("missionId"))
	robotCode := strings.TrimSpace(r.URL.Query().Get("robotCode"))
	latest, err := s.services.Sensors.ListLatestSensorSamples(r.Context(), missionID, robotCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.SensorLatestList(missionID, robotCode, latest))
}

func (s *Server) handleCreateSensorSamples(w http.ResponseWriter, r *http.Request) {
	envelope, err := decodeSensorEnvelope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	samples, err := s.services.Sensors.SaveSensorEnvelope(r.Context(), envelope)
	if err != nil {
		if errors.Is(err, store.ErrInvalidState) || errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.SensorSamplesPayload(samples))
}

func decodeSensorEnvelope(r *http.Request) (domain.SensorEnvelope, error) {
	defer r.Body.Close()
	rawPayload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return domain.SensorEnvelope{}, err
	}
	var request dto.SensorEnvelopeRequest
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
		sensorType, ok := domain.ParseSensorType(descriptor.SensorType)
		if !ok {
			return domain.SensorEnvelope{}, errors.New("descriptor sensorType is required and must be one of: battery, gas, imu, odometry, point_cloud, position")
		}
		envelope.Descriptors = append(envelope.Descriptors, domain.SensorDescriptor{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    sensorID,
			ChannelRole: utils.FirstNonEmptyString(descriptor.ChannelRole, request.ChannelRole),
			Label:       utils.FirstNonEmptyString(descriptor.Label, sensorID),
			SensorType:  string(sensorType),
			Unit:        strings.TrimSpace(descriptor.Unit),
			Enabled:     descriptor.Enabled,
		})
	}
	for _, sample := range request.Samples {
		sensorID := strings.TrimSpace(sample.SensorID)
		if sensorID == "" {
			continue
		}
		if sample.Values == nil && strings.TrimSpace(sample.ObjectKey) == "" {
			return domain.SensorEnvelope{}, errors.New("sample values or objectKey is required")
		}
		envelope.Samples = append(envelope.Samples, domain.SensorSample{
			MissionID:   request.MissionID,
			RobotCode:   request.RobotCode,
			SensorID:    sensorID,
			ChannelRole: utils.FirstNonEmptyString(sample.ChannelRole, request.ChannelRole),
			MessageID:   utils.FirstNonEmptyString(sample.MessageID, request.MessageID),
			Timestamp:   sample.Timestamp,
			ReceivedAt:  envelope.ReceivedAt,
			Values:      marshalSensorSampleValues(sample),
			ObjectKey:   strings.TrimSpace(sample.ObjectKey),
			RawPayload:  utils.RawJSONOrEmpty(sample),
		})
	}
	if len(envelope.Descriptors) == 0 && len(envelope.Samples) == 0 {
		return domain.SensorEnvelope{}, errors.New("descriptors or samples are required")
	}
	return envelope, nil
}

func marshalSensorSampleValues(sample dto.SensorSampleRequest) json.RawMessage {
	if sample.Values != nil {
		return utils.RawJSONOrNil(sample.Values)
	}
	return nil
}
