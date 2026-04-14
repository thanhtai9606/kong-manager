package models

import (
	"time"

	"gorm.io/gorm"
)

// KongCluster is a Kong Admin API endpoint the BFF can proxy to (multi-cluster).
type KongCluster struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	Slug         string         `gorm:"uniqueIndex;size:64;not null" json:"slug"`
	AdminBaseURL string         `gorm:"size:512;not null" json:"admin_base_url"`
	AdminToken   string         `gorm:"size:2048" json:"-"`
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
