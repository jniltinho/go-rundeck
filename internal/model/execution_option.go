package model

import "time"

// ExecutionOption stores the value provided for a JobOption during a specific execution.
type ExecutionOption struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	ExecutionID uint   `gorm:"not null;index"`
	OptionName  string `gorm:"not null;size:100"`
	Value       string `gorm:"type:text"`
	CreatedAt   time.Time

	// Relations
	Execution Execution `gorm:"foreignKey:ExecutionID"`
}
