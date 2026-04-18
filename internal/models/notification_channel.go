package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationChannel stores outbound notification integration (Slack, Teams, Telegram, email).
// Secret holds channel credentials (webhook URL, bot token, SMTP password); never serialized to JSON API.
type NotificationChannel struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Type        string         `gorm:"size:32;not null;index" json:"type"` // slack, teams, telegram, email
	ConfigJSON  string         `gorm:"type:text" json:"-"`                 // JSON object, non-secret settings
	Secret      string         `gorm:"type:text" json:"-"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
