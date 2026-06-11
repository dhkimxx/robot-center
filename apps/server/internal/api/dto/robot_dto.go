package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
)

type RobotResponse struct {
	ID          string                      `json:"id"`
	RobotCode   string                      `json:"robotCode"`
	DisplayName string                      `json:"displayName"`
	ModelName   string                      `json:"modelName,omitempty"`
	Status      domain.RobotConnectionState `json:"status"`
	LastSeenAt  *time.Time                  `json:"lastSeenAt,omitempty"`
	CreatedAt   time.Time                   `json:"createdAt"`
	UpdatedAt   time.Time                   `json:"updatedAt"`
}

type CreateRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

type UpdateRobotRequest struct {
	DisplayName string `json:"displayName"`
	ModelName   string `json:"modelName"`
}

type RobotConnectionInfoResponse struct {
	ServerURL  string `json:"serverUrl"`
	RobotCode  string `json:"robotCode"`
	RobotToken string `json:"robotToken"`
}

type RobotEnvelopeResponse struct {
	Robot RobotResponse `json:"robot"`
}

type RobotsResponse struct {
	Robots []RobotResponse   `json:"robots"`
	Page   ListPageResponse  `json:"page"`
	Query  ListQueryResponse `json:"query"`
}

type CreateRobotResponse struct {
	Robot          RobotResponse               `json:"robot"`
	ConnectionInfo RobotConnectionInfoResponse `json:"connectionInfo"`
}

type RobotConnectionInfoEnvelopeResponse struct {
	ConnectionInfo RobotConnectionInfoResponse `json:"connectionInfo"`
}

func Robot(robot domain.Robot, now time.Time, heartbeatTTL time.Duration) RobotResponse {
	return RobotResponse{
		ID:          robot.ID,
		RobotCode:   robot.RobotCode,
		DisplayName: robot.DisplayName,
		ModelName:   robot.ModelName,
		Status:      robot.ConnectionState(now, heartbeatTTL),
		LastSeenAt:  robot.LastSeenAt,
		CreatedAt:   robot.CreatedAt,
		UpdatedAt:   robot.UpdatedAt,
	}
}

func Robots(robots []domain.Robot, now time.Time, heartbeatTTL time.Duration) []RobotResponse {
	response := make([]RobotResponse, 0, len(robots))
	for _, robot := range robots {
		response = append(response, Robot(robot, now, heartbeatTTL))
	}
	return response
}

func RobotEnvelope(robot domain.Robot, now time.Time, heartbeatTTL time.Duration) RobotEnvelopeResponse {
	return RobotEnvelopeResponse{
		Robot: Robot(robot, now, heartbeatTTL),
	}
}

func RobotsPayload(robots []domain.Robot, now time.Time, heartbeatTTL time.Duration, listMeta ...ListResponseMeta) RobotsResponse {
	response := RobotsResponse{
		Robots: Robots(robots, now, heartbeatTTL),
	}
	if len(listMeta) > 0 {
		response.Page = listMeta[0].Page
		response.Query = listMeta[0].Query
	}
	return response
}

func RobotConnectionInfo(info domain.RobotConnectionInfo) RobotConnectionInfoResponse {
	return RobotConnectionInfoResponse(info)
}

func CreateRobotPayload(robot domain.Robot, info domain.RobotConnectionInfo, now time.Time, heartbeatTTL time.Duration) CreateRobotResponse {
	return CreateRobotResponse{
		Robot:          Robot(robot, now, heartbeatTTL),
		ConnectionInfo: RobotConnectionInfo(info),
	}
}

func RobotConnectionInfoPayload(info domain.RobotConnectionInfo) RobotConnectionInfoEnvelopeResponse {
	return RobotConnectionInfoEnvelopeResponse{
		ConnectionInfo: RobotConnectionInfo(info),
	}
}
