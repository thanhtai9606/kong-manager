package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/rbac"
	"gorm.io/gorm"
)

type putUserGroupsBody struct {
	GroupIDs []uint `json:"group_ids"`
}

// PutUserGroups replaces group membership for a user and syncs Casbin grouping.
func PutUserGroups(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uid, err := parseUintParam(r, "userID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body putUserGroupsBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var u models.User
		if err := db.Preload("Groups").First(&u, uid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		var groups []models.Group
		if len(body.GroupIDs) > 0 {
			if err := db.Where("id IN ?", body.GroupIDs).Find(&groups).Error; err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			if len(groups) != len(body.GroupIDs) {
				http.Error(w, "unknown group id", http.StatusBadRequest)
				return
			}
		}
		if err := db.Model(&u).Association("Groups").Replace(groups); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		names := make([]string, 0, len(groups))
		for _, g := range groups {
			names = append(names, g.Name)
		}
		if err := rbac.SyncUserGroups(e, u.Username, names); err != nil {
			http.Error(w, "casbin sync error", http.StatusInternalServerError)
			return
		}
		if err := db.Preload("Groups").First(&u, u.ID).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	}
}
