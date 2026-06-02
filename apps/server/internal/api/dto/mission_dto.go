package dto

import (
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/store"
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

type CreateMissionRequest struct {
	Name        string   `json:"name"`
	MissionType string   `json:"missionType"`
	SiteNote    string   `json:"siteNote"`
	RobotCode   string   `json:"robotCode"`
	RobotCodes  []string `json:"robotCodes"`
}

type MissionEnvelopeResponse struct {
	Mission MissionResponse `json:"mission"`
}

type MissionsResponse struct {
	Missions []MissionResponse `json:"missions"`
}

type MissionConflictResponse struct {
	RobotCode         string `json:"robotCode"`
	ActiveMissionCode string `json:"activeMissionCode"`
}

type MissionConflictEnvelopeResponse struct {
	Error     string                    `json:"error"`
	Conflicts []MissionConflictResponse `json:"conflicts"`
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

func MissionPayload(mission domain.Mission) MissionEnvelopeResponse {
	return MissionEnvelopeResponse{
		Mission: Mission(mission),
	}
}

func Missions(missions []domain.Mission) []MissionResponse {
	response := make([]MissionResponse, 0, len(missions))
	for _, mission := range missions {
		response = append(response, Mission(mission))
	}
	return response
}

func MissionsPayload(missions []domain.Mission) MissionsResponse {
	return MissionsResponse{
		Missions: Missions(missions),
	}
}

func MissionConflictPayload(errorMessage string, conflicts []store.MissionStartConflict) MissionConflictEnvelopeResponse {
	response := MissionConflictEnvelopeResponse{
		Error:     errorMessage,
		Conflicts: make([]MissionConflictResponse, 0, len(conflicts)),
	}
	for _, conflict := range conflicts {
		response.Conflicts = append(response.Conflicts, MissionConflictResponse{
			RobotCode:         conflict.RobotCode,
			ActiveMissionCode: conflict.ActiveMissionCode,
		})
	}
	return response
}
