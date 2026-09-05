package model

import "time"

type UserRole string

const (
	RoleSuperadmin UserRole = "superadmin"
	RoleAdmin      UserRole = "admin"
	RoleViewer     UserRole = "viewer"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        *string   `gorm:"size:255" json:"email,omitempty"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         UserRole  `gorm:"size:20;not null;default:viewer" json:"role"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}

func (User) TableName() string { return "users" }
