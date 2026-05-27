package model

import (
	"time"

	"robot-center/apps/server/internal/domain"
	"robot-center/apps/server/internal/utils"
)

type MissionModel struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionCode string     `gorm:"column:mission_code;not null;uniqueIndex"`
	Name        string     `gorm:"column:name;not null"`
	MissionType string     `gorm:"column:mission_type;not null"`
	Status      string     `gorm:"column:status;not null;default:ready"`
	CreatedBy   *string    `gorm:"column:created_by;type:uuid;index"`
	SiteNote    *string    `gorm:"column:site_note"`
	StartedAt   *time.Time `gorm:"column:started_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (MissionModel) TableName() string {
	return "missions"
}

func (record MissionModel) ToDomainMission(robotCodes []string) domain.Mission {
	robotCodes = append([]string(nil), robotCodes...)
	return domain.Mission{
		ID:          record.ID,
		MissionCode: record.MissionCode,
		Name:        record.Name,
		MissionType: record.MissionType,
		Status:      record.Status,
		SiteNote:    stringFromPointer(record.SiteNote),
		RobotCode:   utils.FirstString(robotCodes),
		RobotCodes:  robotCodes,
		StartedAt:   record.StartedAt,
		EndedAt:     record.EndedAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

type MissionRobotModel struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	MissionID string     `gorm:"column:mission_id;type:uuid;not null;index"`
	RobotID   string     `gorm:"column:robot_id;type:uuid;not null;index"`
	Role      string     `gorm:"column:role;not null;default:primary"`
	Status    string     `gorm:"column:status;not null;default:assigned"`
	JoinedAt  *time.Time `gorm:"column:joined_at"`
	LeftAt    *time.Time `gorm:"column:left_at"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (MissionRobotModel) TableName() string {
	return "mission_robots"
}
