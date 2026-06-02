package dto

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type SensorDescriptorResponse struct {
	ID          string    `json:"id"`
	MissionID   string    `json:"missionId"`
	RobotCode   string    `json:"robotCode"`
	SensorID    string    `json:"sensorId"`
	ChannelRole string    `json:"channelRole"`
	Label       string    `json:"label"`
	SensorType  string    `json:"sensorType"`
	Unit        string    `json:"unit,omitempty"`
	Enabled     bool      `json:"enabled"`
	FirstSeenAt time.Time `json:"firstSeenAt"`
	LastSeenAt  time.Time `json:"lastSeenAt"`
}

type SensorDescriptorRequest struct {
	SensorID    string `json:"sensorId"`
	ChannelRole string `json:"channelRole"`
	Label       string `json:"label"`
	SensorType  string `json:"sensorType"`
	Unit        string `json:"unit"`
	Enabled     bool   `json:"enabled"`
}

type SensorSampleRequest struct {
	SensorID    string     `json:"sensorId,omitempty"`
	ChannelRole string     `json:"channelRole,omitempty"`
	MessageID   string     `json:"messageId,omitempty"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Values      any        `json:"values,omitempty"`
	ObjectKey   string     `json:"objectKey,omitempty"`
}

type SensorEnvelopeRequest struct {
	MessageID   string                    `json:"messageId"`
	MessageType string                    `json:"messageType"`
	RobotCode   string                    `json:"robotCode"`
	MissionID   string                    `json:"missionId"`
	ChannelRole string                    `json:"channelRole"`
	Descriptors []SensorDescriptorRequest `json:"descriptors"`
	Samples     []SensorSampleRequest     `json:"samples"`
}

type SensorSampleResponse struct {
	ID           string          `json:"id"`
	DescriptorID string          `json:"descriptorId,omitempty"`
	MissionID    string          `json:"missionId"`
	RobotCode    string          `json:"robotCode"`
	SensorID     string          `json:"sensorId"`
	ChannelRole  string          `json:"channelRole"`
	MessageID    string          `json:"messageId,omitempty"`
	Timestamp    *time.Time      `json:"timestamp,omitempty"`
	ReceivedAt   time.Time       `json:"receivedAt"`
	Values       json.RawMessage `json:"values,omitempty"`
	ObjectKey    string          `json:"objectKey,omitempty"`
}

type SensorLatestResponse struct {
	SensorDescriptorResponse
	LatestSample *SensorSampleResponse        `json:"latestSample,omitempty"`
	Readings     []SensorValueReadingResponse `json:"readings,omitempty"`
}

type SensorValueReadingResponse struct {
	FieldPath string  `json:"fieldPath"`
	Label     string  `json:"label"`
	Order     float64 `json:"order"`
	Unit      string  `json:"unit,omitempty"`
	Value     any     `json:"value"`
}

type SensorDescriptorsResponse struct {
	SensorDescriptors []SensorDescriptorResponse `json:"sensorDescriptors"`
}

type SensorSamplesResponse struct {
	SensorSamples []SensorSampleResponse `json:"sensorSamples"`
}

type SensorLatestListResponse struct {
	MissionID string                 `json:"missionId"`
	RobotCode string                 `json:"robotCode"`
	Sensors   []SensorLatestResponse `json:"sensors"`
}

func SensorDescriptor(descriptor domain.SensorDescriptor) SensorDescriptorResponse {
	return SensorDescriptorResponse{
		ID:          descriptor.ID,
		MissionID:   descriptor.MissionID,
		RobotCode:   descriptor.RobotCode,
		SensorID:    descriptor.SensorID,
		ChannelRole: descriptor.ChannelRole,
		Label:       descriptor.Label,
		SensorType:  descriptor.SensorType,
		Unit:        descriptor.Unit,
		Enabled:     descriptor.Enabled,
		FirstSeenAt: descriptor.FirstSeenAt,
		LastSeenAt:  descriptor.LastSeenAt,
	}
}

func SensorDescriptors(descriptors []domain.SensorDescriptor) []SensorDescriptorResponse {
	response := make([]SensorDescriptorResponse, 0, len(descriptors))
	for _, descriptor := range descriptors {
		response = append(response, SensorDescriptor(descriptor))
	}
	return response
}

func SensorDescriptorsPayload(descriptors []domain.SensorDescriptor) SensorDescriptorsResponse {
	return SensorDescriptorsResponse{
		SensorDescriptors: SensorDescriptors(descriptors),
	}
}

func SensorSample(sample domain.SensorSample) SensorSampleResponse {
	return SensorSampleResponse{
		ID:           sample.ID,
		DescriptorID: sample.DescriptorID,
		MissionID:    sample.MissionID,
		RobotCode:    sample.RobotCode,
		SensorID:     sample.SensorID,
		ChannelRole:  sample.ChannelRole,
		MessageID:    sample.MessageID,
		Timestamp:    sample.Timestamp,
		ReceivedAt:   sample.ReceivedAt,
		Values:       sample.Values,
		ObjectKey:    sample.ObjectKey,
	}
}

func SensorSamples(samples []domain.SensorSample) []SensorSampleResponse {
	response := make([]SensorSampleResponse, 0, len(samples))
	for _, sample := range samples {
		response = append(response, SensorSample(sample))
	}
	return response
}

func SensorSamplesPayload(samples []domain.SensorSample) SensorSamplesResponse {
	return SensorSamplesResponse{
		SensorSamples: SensorSamples(samples),
	}
}

func SensorLatest(items []domain.SensorLatest) []SensorLatestResponse {
	response := make([]SensorLatestResponse, 0, len(items))
	for _, item := range items {
		latest := SensorLatestResponse{
			SensorDescriptorResponse: SensorDescriptor(item.Descriptor),
		}
		if item.LatestSample != nil {
			sample := SensorSample(*item.LatestSample)
			latest.LatestSample = &sample
			latest.Readings = SensorValueReadings(domain.InterpretSensorSampleValue(item.Descriptor, *item.LatestSample))
		}
		response = append(response, latest)
	}
	return response
}

func SensorLatestList(missionID string, robotCode string, items []domain.SensorLatest) SensorLatestListResponse {
	return SensorLatestListResponse{
		MissionID: missionID,
		RobotCode: robotCode,
		Sensors:   SensorLatest(items),
	}
}

func SensorValueReadings(readings []domain.SensorValueReading) []SensorValueReadingResponse {
	response := make([]SensorValueReadingResponse, 0, len(readings))
	for _, reading := range readings {
		response = append(response, SensorValueReadingResponse{
			FieldPath: reading.FieldPath,
			Label:     reading.Label,
			Order:     reading.Order,
			Unit:      reading.Unit,
			Value:     reading.Value,
		})
	}
	return response
}
