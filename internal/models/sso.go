package models

import (
	"time"

	"gorm.io/gorm"
)

// SSOProvider configures an OpenID Connect identity provider (Keycloak, Azure AD, etc.).
type SSOProvider struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Slug         string         `gorm:"uniqueIndex;size:64;not null" json:"slug"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	IssuerURL    string         `gorm:"size:512;not null" json:"issuer_url"`
	ClientID     string         `gorm:"size:256;not null" json:"client_id"`
	ClientSecret string         `gorm:"size:512;not null" json:"-"`
	Scopes       string         `gorm:"size:512" json:"scopes"`
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
