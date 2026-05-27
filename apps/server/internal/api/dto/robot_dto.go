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

type RobotConnectionInfoResponse struct {
	ServerURL  string `json:"serverUrl"`
	RobotCode  string `json:"robotCode"`
	RobotToken string `json:"robotToken"`
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

func RobotConnectionInfo(info domain.RobotConnectionInfo) RobotConnectionInfoResponse {
	return RobotConnectionInfoResponse(info)
}
