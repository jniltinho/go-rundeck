package model

import "gorm.io/gorm"

// AuthType defines how the node authenticates SSH connections.
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"
)

// Node represents a remote host that can execute job steps.
type Node struct {
	gorm.Model
	ProjectID   uint     `gorm:"not null;index"`
	Name        string   `gorm:"not null;size:150"`
	Hostname    string   `gorm:"not null;size:255"`
	SSHPort     int      `gorm:"not null;default:22"`
	SSHUser     string   `gorm:"not null;size:100"`
	AuthType    AuthType `gorm:"not null;default:'key'"`
	KeyID       *uint    `gorm:"index"`
	Tags        string   `gorm:"size:255"`
	Description string   `gorm:"size:500"`
	OSFamily    string   `gorm:"size:50"`
	Active      bool     `gorm:"not null;default:true"`

	// Relations
	Project    Project     `gorm:"foreignKey:ProjectID"`
	KeyStorage *KeyStorage `gorm:"foreignKey:KeyID"`
}
