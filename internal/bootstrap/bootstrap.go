package bootstrap

import (
	"log"
	"strings"

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

	if err := seedDefaultKongCluster(cfg, db); err != nil {
		return err
	}

	var n int64
	if err := db.Model(&models.User{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return syncAllCasbin(cfg, db, enforcer)
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

// seedDefaultKongCluster inserts the "default" row once; KONG_ADMIN_URL is only used here.
// Runtime routing uses kong_clusters rows (see proxy.DynamicKongHandler).
func seedDefaultKongCluster(cfg *config.Config, db *gorm.DB) error {
	var n int64
	if err := db.Model(&models.KongCluster{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	base := strings.TrimSpace(cfg.KongAdminURL)
	base = strings.TrimRight(base, "/")
	if base == "" {
		base = "http://127.0.0.1:8001"
	}
	return db.Create(&models.KongCluster{
		Name:         "Default",
		Slug:         "default",
		AdminBaseURL: base,
		Enabled:      true,
	}).Error
}

func syncAllCasbin(cfg *config.Config, db *gorm.DB, enforcer *casbin.Enforcer) error {
	var users []models.User
	if err := db.Preload("Groups").Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		names := make([]string, 0, len(u.Groups))
		for _, g := range u.Groups {
			names = append(names, g.Name)
		}
		if len(names) == 0 {
			var grp models.Group
			target := "viewer"
			if u.Username == "admin" || (cfg.BootstrapUsername != "" && u.Username == cfg.BootstrapUsername) {
				target = "admin"
			}
			if err := db.Where("name = ?", target).First(&grp).Error; err != nil {
				log.Printf("bootstrap: user %q has no groups; cannot assign %q: %v", u.Username, target, err)
				continue
			}
			if err := db.Model(&u).Association("Groups").Append(&grp); err != nil {
				log.Printf("bootstrap: user %q group repair: %v", u.Username, err)
				continue
			}
			names = []string{target}
			log.Printf("bootstrap: repaired user %q → group %q (was missing groups)", u.Username, target)
		}
		if err := rbac.SyncUserGroups(enforcer, u.Username, names); err != nil {
			return err
		}
	}
	return nil
}
