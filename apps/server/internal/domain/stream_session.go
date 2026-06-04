package domain

import "time"

type RobotStreamSession struct {
	ID              string
	MissionID       string
	MissionCode     string
	RobotID         string
	RobotCode       string
	PublisherPeerID string
	State           string
	StartedAt       time.Time
	LastMediaAt     *time.Time
	EndedAt         *time.Time
	EndReason       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
