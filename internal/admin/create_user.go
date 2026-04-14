package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
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

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.@-]{1,191}$`)

type createUserBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	GroupIDs []uint `json:"group_ids"`
}

// CreateUser adds a local account and optional group membership (admin API).
func CreateUser(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		claims, ok := httpapi.ClaimsFrom(r.Context())
		if !ok || claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body createUserBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		body.Username = strings.TrimSpace(body.Username)
		if !usernameRe.MatchString(body.Username) {
			http.Error(w, "invalid username (1–191 chars: letters, digits, _, ., @, -)", http.StatusBadRequest)
			return
		}
		if len(body.Password) < 8 {
			http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
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
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "hash error", http.StatusInternalServerError)
			return
		}
		u := models.User{
			Username:     body.Username,
			PasswordHash: string(hash),
		}
		if err := db.Create(&u).Error; err != nil {
			http.Error(w, "could not create user (duplicate username?)", http.StatusConflict)
			return
		}
		if len(groups) > 0 {
			if err := db.Model(&u).Association("Groups").Append(groups); err != nil {
				_ = db.Unscoped().Delete(&models.User{}, u.ID)
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
		}
		names := make([]string, 0, len(groups))
		for _, g := range groups {
			names = append(names, g.Name)
		}
		if err := rbac.SyncUserGroups(e, u.Username, names); err != nil {
			_ = db.Unscoped().Delete(&models.User{}, u.ID)
			http.Error(w, "casbin sync error", http.StatusInternalServerError)
			return
		}
		if err := db.Preload("Groups").First(&u, u.ID).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		details, _ := json.Marshal(map[string]any{"group_ids": body.GroupIDs})
		if err := audit.Append(db, r, claims.Username, "user.create", "user:"+strconv.FormatUint(uint64(u.ID), 10), string(details)); err != nil {
			log.Printf("audit: user.create: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(u)
	}
}
