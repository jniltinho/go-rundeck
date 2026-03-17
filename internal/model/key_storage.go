package model

import "time"

// KeyType classifies the credential stored.
type KeyType string

const (
	KeyTypePrivateKey KeyType = "private_key"
	KeyTypePassword   KeyType = "password"
)

// KeyStorage holds encrypted SSH keys and passwords.
type KeyStorage struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	ProjectID   *uint     `gorm:"index"`
	Name        string    `gorm:"not null;size:150"`
	KeyType     KeyType   `gorm:"not null"`
	ContentEnc  string    `gorm:"not null;type:text"` // AES-256-GCM encrypted, base64 encoded
	Description string    `gorm:"size:500"`
	CreatedBy   uint      `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations
	Project *Project `gorm:"foreignKey:ProjectID"`
}
