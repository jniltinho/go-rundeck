package model

import "gorm.io/gorm"

// ExecStrategy controls how steps run across nodes.
type ExecStrategy string

const (
	ExecStrategySequential ExecStrategy = "sequential"
	ExecStrategyParallel   ExecStrategy = "parallel"
)

// OnError controls execution behaviour when a step fails.
type OnError string

const (
	OnErrorStop     OnError = "stop"
	OnErrorContinue OnError = "continue"
)

// Job defines a runbook: a list of steps to execute on filtered nodes.
type Job struct {
	gorm.Model
	ProjectID     uint         `gorm:"not null;index"`
	Name          string       `gorm:"not null;size:150"`
	Description   string       `gorm:"size:500"`
	NodeFilter    string       `gorm:"size:255"` // tag or name expression
	ExecStrategy  ExecStrategy `gorm:"not null;default:'sequential'"`
	OnError       OnError      `gorm:"not null;default:'stop'"`
	TimeoutSec    int          `gorm:"default:300"`
	CreatedBy     uint         `gorm:"not null"`

	// Relations
	Project    Project     `gorm:"foreignKey:ProjectID"`
	Steps      []JobStep   `gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE"`
	Executions []Execution `gorm:"foreignKey:JobID"`
	Schedules  []Schedule  `gorm:"foreignKey:JobID"`
}
