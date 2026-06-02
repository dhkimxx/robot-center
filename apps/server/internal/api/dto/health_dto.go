package dto

import "time"

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	StartedAt string `json:"startedAt"`
}

func Health(startedAt time.Time) HealthResponse {
	return HealthResponse{
		Status:    "ok",
		Service:   "app-server",
		StartedAt: startedAt.Format(time.RFC3339),
	}
}
