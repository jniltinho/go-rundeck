package model

import "gorm.io/gorm"

// Role represents a user permission level.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// User represents an authenticated user of the platform.
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null;size:100"`
	PasswordHash string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;size:255"`
	Role         Role   `gorm:"not null;default:'viewer'"`
	Active       bool   `gorm:"not null;default:true"`
}
