package model

import "time"

// LogLevel classifies a log entry severity.
type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelDebug LogLevel = "debug"
)

// ExecutionLog stores a single line of output from an execution.
type ExecutionLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	ExecutionID uint      `gorm:"not null;index"`
	NodeName    string    `gorm:"size:150"`
	StepOrder   int       `gorm:"not null"`
	LogLevel    LogLevel  `gorm:"not null;default:'info'"`
	Message     string    `gorm:"not null;type:text"`
	LoggedAt    time.Time `gorm:"not null"`

	// Relations
	Execution Execution `gorm:"foreignKey:ExecutionID"`
}
