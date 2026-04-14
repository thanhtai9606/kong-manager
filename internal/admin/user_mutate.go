package admin

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/kong/kong-manager/internal/audit"
	"github.com/kong/kong-manager/internal/httpapi"
	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/rbac"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type patchUserBody struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

// PatchUser updates username and/or password (admin API).
func PatchUser(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		claims, ok := httpapi.ClaimsFrom(r.Context())
		if !ok || claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		uid, err := parseUintParam(r, "userID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body patchUserBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if body.Username == nil && body.Password == nil {
			http.Error(w, "username or password required", http.StatusBadRequest)
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
		oldName := u.Username
		changed := false
		if body.Username != nil {
			s := strings.TrimSpace(*body.Username)
			if !usernameRe.MatchString(s) {
				http.Error(w, "invalid username (letters, digits, _, ., @, -)", http.StatusBadRequest)
				return
			}
			if s != u.Username {
				var n int64
				if err := db.Model(&models.User{}).Where("username = ? AND id <> ?", s, u.ID).Count(&n).Error; err != nil {
					http.Error(w, "database error", http.StatusInternalServerError)
					return
				}
				if n > 0 {
					http.Error(w, "username already taken", http.StatusConflict)
					return
				}
				u.Username = s
				changed = true
			}
		}
		if body.Password != nil {
			p := *body.Password
			if len(p) < 8 {
				http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
				return
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "hash error", http.StatusInternalServerError)
				return
			}
			u.PasswordHash = string(hash)
			changed = true
		}
		if !changed {
			if err := db.Preload("Groups").First(&u, u.ID).Error; err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(u)
			return
		}
		if err := db.Save(&u).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		names := make([]string, 0, len(u.Groups))
		for _, g := range u.Groups {
			names = append(names, g.Name)
		}
		if oldName != u.Username {
			if err := rbac.SyncUserGroups(e, oldName, nil); err != nil {
				http.Error(w, "casbin sync error", http.StatusInternalServerError)
				return
			}
			if err := rbac.SyncUserGroups(e, u.Username, names); err != nil {
				http.Error(w, "casbin sync error", http.StatusInternalServerError)
				return
			}
		}
		if err := db.Preload("Groups").First(&u, u.ID).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		details, _ := json.Marshal(map[string]any{
			"username_changed": body.Username != nil && oldName != u.Username,
			"password_changed": body.Password != nil,
		})
		if err := audit.Append(db, r, claims.Username, "user.update", "user:"+strconv.FormatUint(uint64(u.ID), 10), string(details)); err != nil {
			log.Printf("audit: user.update: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	}
}

// DeleteUser removes a local user and Casbin grouping.
func DeleteUser(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		claims, ok := httpapi.ClaimsFrom(r.Context())
		if !ok || claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		uid, err := parseUintParam(r, "userID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var u models.User
		if err := db.First(&u, uid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if claims.Username == u.Username {
			http.Error(w, "cannot delete your own account", http.StatusBadRequest)
			return
		}
		if err := db.Model(&u).Association("Groups").Clear(); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if err := rbac.SyncUserGroups(e, u.Username, nil); err != nil {
			http.Error(w, "casbin sync error", http.StatusInternalServerError)
			return
		}
		if err := db.Unscoped().Delete(&models.User{}, uid).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if err := audit.Append(db, r, claims.Username, "user.delete", "user:"+strconv.FormatUint(uid, 10), "{}"); err != nil {
			log.Printf("audit: user.delete: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
