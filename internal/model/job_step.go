package model

import "time"

// StepType classifies what kind of command a step runs.
type StepType string

const (
	StepTypeCommand StepType = "command"
	StepTypeScript  StepType = "script"
)

// JobStep is a single unit of work within a Job.
type JobStep struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	JobID       uint      `gorm:"not null;index"`
	StepOrder   int       `gorm:"not null"`
	Type        StepType  `gorm:"not null;default:'command'"`
	Label       string    `gorm:"size:150"`
	Content     string    `gorm:"not null;type:text"`
	Interpreter string    `gorm:"size:100"` // e.g. /bin/bash
	Args        string    `gorm:"size:500"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations
	Job Job `gorm:"foreignKey:JobID"`
}
