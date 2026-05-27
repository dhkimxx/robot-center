package domain

import (
	"strings"
	"time"
)

type Mission struct {
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

func StreamRoomID(missionCode string, robotCode string) string {
	missionCode = strings.TrimSpace(missionCode)
	robotCode = strings.TrimSpace(robotCode)
	if missionCode == "" || robotCode == "" {
		return missionCode
	}
	return missionCode + "__" + robotCode
}
