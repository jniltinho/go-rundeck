package model

import "time"

// Schedule defines a cron-based trigger for a Job.
type Schedule struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	JobID     uint       `gorm:"not null;index"`
	CronExpr  string     `gorm:"not null;size:100"`
	Enabled   bool       `gorm:"not null;default:true"`
	NextRun   *time.Time
	LastRun   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Job Job `gorm:"foreignKey:JobID"`
}
