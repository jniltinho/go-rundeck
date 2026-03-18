package model

import "time"

// OptionType defines what kind of input the user provides.
type OptionType string

const (
	OptionTypeText   OptionType = "text"
	OptionTypeSecure OptionType = "secure"
	OptionTypeChoice OptionType = "choice"
)

// JobOption represents a variable that can be provided when running a job.
type JobOption struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	JobID       uint       `gorm:"not null;index"`
	Name        string     `gorm:"not null;size:100"` // Variable name (e.g. "env")
	Label       string     `gorm:"size:150"`          // Display label
	Description string     `gorm:"size:500"`
	OptionType  OptionType `gorm:"not null;default:'text'"`
	Required    bool       `gorm:"not null;default:false"`
	DefaultVal  string     `gorm:"size:255"`          // Default value if not provided
	Choices     string     `gorm:"size:500"`          // Comma-separated if type is 'choice'
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations
	Job Job `gorm:"foreignKey:JobID"`
}
