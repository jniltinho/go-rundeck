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
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	JobID       uint      `gorm:"not null;index"           json:"job_id,omitempty"`
	StepOrder   int       `gorm:"not null"                 json:"step_order"`
	Type        StepType  `gorm:"not null;default:'command'" json:"type"`
	Label       string    `gorm:"size:150"                 json:"label"`
	Content     string    `gorm:"not null;type:text"       json:"content"`
	Interpreter string    `gorm:"size:100"                 json:"interpreter"`
	Args        string    `gorm:"size:500"                 json:"args"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	// Relations
	Job Job `gorm:"foreignKey:JobID" json:"-"`
}
