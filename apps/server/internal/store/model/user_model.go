package model

import "time"

type UserModel struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	LoginID      string     `gorm:"column:login_id;not null;uniqueIndex"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	DisplayName  string     `gorm:"column:display_name;not null"`
	Role         string     `gorm:"column:role;not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (UserModel) TableName() string {
	return "users"
}
