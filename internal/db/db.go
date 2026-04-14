package db

import (
	"fmt"

	"github.com/kong/kong-manager/internal/config"
	"github.com/kong/kong-manager/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open returns a GORM DB connection based on DATABASE_DRIVER / DATABASE_URL.
func Open(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.DatabaseDriver {
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DatabaseURL)
	case "mysql":
		dialector = mysql.Open(cfg.DatabaseURL)
	case "sqlite", "sqlite3":
		dialector = sqlite.Open(cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unsupported DATABASE_DRIVER: %q (use postgres, mysql, sqlite)", cfg.DatabaseDriver)
	}

	return gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

// AutoMigrate runs schema migrations for app models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.KongCluster{},
		&models.AuditLog{},
		&models.SSOProvider{},
	)
}
