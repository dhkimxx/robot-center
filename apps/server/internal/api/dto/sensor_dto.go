package dto

import (
	"encoding/json"
	"time"

	"robot-center/apps/server/internal/domain"
)

type SensorDescriptorResponse struct {
	ID          string          `json:"id"`
	MissionID   string          `json:"missionId"`
	RobotCode   string          `json:"robotCode"`
	SensorID    string          `json:"sensorId"`
	ChannelRole string          `json:"channelRole"`
	DisplayName string          `json:"displayName"`
	SensorType  string          `json:"sensorType"`
	Unit        string          `json:"unit,omitempty"`
	Enabled     bool            `json:"enabled"`
	Metadata    json.RawMessage `json:"metadata"`
	FirstSeenAt time.Time       `json:"firstSeenAt"`
	LastSeenAt  time.Time       `json:"lastSeenAt"`
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
	LatestSample *SensorSampleResponse `json:"latestSample,omitempty"`
}

func SensorDescriptor(descriptor domain.SensorDescriptor) SensorDescriptorResponse {
	return SensorDescriptorResponse(descriptor)
}

func SensorDescriptors(descriptors []domain.SensorDescriptor) []SensorDescriptorResponse {
	response := make([]SensorDescriptorResponse, 0, len(descriptors))
	for _, descriptor := range descriptors {
		response = append(response, SensorDescriptor(descriptor))
	}
	return response
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

func SensorLatest(items []domain.SensorLatest) []SensorLatestResponse {
	response := make([]SensorLatestResponse, 0, len(items))
	for _, item := range items {
		latest := SensorLatestResponse{
			SensorDescriptorResponse: SensorDescriptor(item.Descriptor),
		}
		if item.LatestSample != nil {
			sample := SensorSample(*item.LatestSample)
			latest.LatestSample = &sample
		}
		response = append(response, latest)
	}
	return response
}
