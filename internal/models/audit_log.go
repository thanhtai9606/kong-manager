package models

import (
	"time"
)

// AuditLog records privileged actions in the BFF (append-only).
type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ActorUsername string    `gorm:"size:191;not null;index" json:"actor_username"`
	Action        string    `gorm:"size:128;not null;index" json:"action"`
	Resource      string    `gorm:"size:256" json:"resource"`
	Details       string    `gorm:"type:text" json:"details,omitempty"`
	ClientIP      string    `gorm:"size:64" json:"client_ip,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
