package model

import "time"

type UserModel struct {
	BaseModel
	LoginID      string     `gorm:"column:login_id;not null;uniqueIndex"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	DisplayName  string     `gorm:"column:display_name;not null"`
	Role         string     `gorm:"column:role;not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
}

func (UserModel) TableName() string {
	return "users"
}
