package model

import "gorm.io/gorm"

// Project groups nodes and jobs under a single namespace.
type Project struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:150"`
	Description string `gorm:"size:500"`
	Tags        string `gorm:"size:255"`
	Active      bool   `gorm:"not null;default:true"`
	CreatedBy   uint   `gorm:"not null"`

	// Relations
	Nodes      []Node      `gorm:"foreignKey:ProjectID"`
	Jobs       []Job       `gorm:"foreignKey:ProjectID"`
	Executions []Execution `gorm:"foreignKey:ProjectID"`
}
