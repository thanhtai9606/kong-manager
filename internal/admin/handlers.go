package admin

import (
	"encoding/json"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

// ListUsers returns all local users with their groups (no password fields).
func ListUsers(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var users []models.User
		if err := db.Preload("Groups").Order("username").Find(&users).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	}
}

// RBACSnapshot returns Casbin policy rows (p) and grouping rows (g) for inspection.
func RBACSnapshot(e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		policies, err := e.GetPolicy()
		if err != nil {
			http.Error(w, "casbin policy read", http.StatusInternalServerError)
			return
		}
		grouping, err := e.GetGroupingPolicy()
		if err != nil {
			http.Error(w, "casbin grouping read", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"policies": policies,
			"grouping": grouping,
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	}
}
