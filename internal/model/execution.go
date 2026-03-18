package model

import "time"

// ExecutionStatus represents the lifecycle state of an execution.
type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSucceeded ExecutionStatus = "succeeded"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusAborted   ExecutionStatus = "aborted"
)

// TriggerType identifies what initiated the execution.
type TriggerType string

const (
	TriggerTypeManual   TriggerType = "manual"
	TriggerTypeSchedule TriggerType = "schedule"
)

// Execution records a single run of a Job.
type Execution struct {
	ID          uint            `gorm:"primaryKey;autoIncrement"`
	JobID       uint            `gorm:"not null;index"`
	ProjectID   uint            `gorm:"not null;index"`
	Status      ExecutionStatus `gorm:"not null;default:'running'"`
	TriggeredBy *uint           `gorm:"index"`
	TriggerType TriggerType     `gorm:"not null;default:'manual'"`
	StartedAt   time.Time       `gorm:"not null"`
	EndedAt     *time.Time
	DurationSec *float64
	CreatedAt   time.Time

	// Relations
	Job     Job               `gorm:"foreignKey:JobID"`
	Options []ExecutionOption `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
	Logs    []ExecutionLog    `gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
}
