package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
)

type MissionResponse struct {
	ID          string     `json:"id"`
	MissionCode string     `json:"missionCode"`
	Name        string     `json:"name"`
	MissionType string     `json:"missionType"`
	Status      string     `json:"status"`
	SiteNote    string     `json:"siteNote,omitempty"`
	RobotCode   string     `json:"robotCode,omitempty"`
	RobotCodes  []string   `json:"robotCodes,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func Mission(mission domain.Mission) MissionResponse {
	return MissionResponse{
		ID:          mission.ID,
		MissionCode: mission.MissionCode,
		Name:        mission.Name,
		MissionType: mission.MissionType,
		Status:      mission.Status,
		SiteNote:    mission.SiteNote,
		RobotCode:   mission.RobotCode,
		RobotCodes:  append([]string(nil), mission.RobotCodes...),
		StartedAt:   mission.StartedAt,
		EndedAt:     mission.EndedAt,
		CreatedAt:   mission.CreatedAt,
		UpdatedAt:   mission.UpdatedAt,
	}
}

func Missions(missions []domain.Mission) []MissionResponse {
	response := make([]MissionResponse, 0, len(missions))
	for _, mission := range missions {
		response = append(response, Mission(mission))
	}
	return response
}
