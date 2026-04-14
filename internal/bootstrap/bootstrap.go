package bootstrap

import (
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/kong/kong-manager/internal/config"
	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/rbac"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Run creates default groups, optional bootstrap admin, and Casbin policies.
func Run(cfg *config.Config, db *gorm.DB, enforcer *casbin.Enforcer) error {
	if err := rbac.SeedDefaultPolicies(enforcer); err != nil {
		return err
	}

	var adminGroup models.Group
	if err := db.Where("name = ?", "admin").FirstOrCreate(&adminGroup, models.Group{Name: "admin"}).Error; err != nil {
		return err
	}
	var viewerGroup models.Group
	if err := db.Where("name = ?", "viewer").FirstOrCreate(&viewerGroup, models.Group{Name: "viewer"}).Error; err != nil {
		return err
	}

	var n int64
	if err := db.Model(&models.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return syncAllCasbin(db, enforcer)
	}

	if cfg.BootstrapUsername == "" || cfg.BootstrapPassword == "" {
		log.Print("bootstrap: no users and BOOTSTRAP_ADMIN_USERNAME/PASSWORD unset — create first user via DB or set env")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.BootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := models.User{
		Username:     cfg.BootstrapUsername,
		PasswordHash: string(hash),
		Groups:       []models.Group{adminGroup},
	}
	if err := db.Create(&u).Error; err != nil {
		return err
	}
	if err := rbac.SyncUserGroups(enforcer, u.Username, []string{"admin"}); err != nil {
		return err
	}
	log.Printf("bootstrap: created admin user %q", u.Username)
	return nil
}

func syncAllCasbin(db *gorm.DB, enforcer *casbin.Enforcer) error {
	var users []models.User
	if err := db.Preload("Groups").Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		names := make([]string, 0, len(u.Groups))
		for _, g := range u.Groups {
			names = append(names, g.Name)
		}
		if err := rbac.SyncUserGroups(enforcer, u.Username, names); err != nil {
			return err
		}
	}
	return nil
}
