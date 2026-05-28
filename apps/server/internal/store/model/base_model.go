package model

import "time"

type BaseModel struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now();autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now();autoUpdateTime"`
}
