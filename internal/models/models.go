package models

import (
	"time"

	"gorm.io/gorm"
)

// User is a local account for Kong Manager.
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:191;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Email        string         `gorm:"size:191" json:"email,omitempty"`
	SSOProviderID *uint         `gorm:"index" json:"sso_provider_id,omitempty"`
	ExternalSub  string         `gorm:"size:512" json:"external_sub,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Groups       []Group        `gorm:"many2many:user_auth_groups;" json:"groups,omitempty"`
}

// Group is a named role bucket for Casbin grouping (e.g. admin, viewer).
type Group struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;size:64;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Users     []User         `gorm:"many2many:user_auth_groups;" json:"-"`
}

func (Group) TableName() string {
	return "auth_groups"
}
