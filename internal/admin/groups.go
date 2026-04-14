package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/rbac"
	"gorm.io/gorm"
)

var groupNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func parseUintParam(r *http.Request, key string) (uint64, error) {
	return strconv.ParseUint(chi.URLParam(r, key), 10, 64)
}

// ListGroups returns all roles (auth_groups).
func ListGroups(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var groups []models.Group
		if err := db.Order("name").Find(&groups).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(groups)
	}
}

type createGroupBody struct {
	Name string `json:"name"`
}

// CreateGroup adds a role with no Casbin policies until configured.
func CreateGroup(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body createGroupBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		name := body.Name
		if !groupNameRe.MatchString(name) {
			http.Error(w, "invalid name (use 1-64 chars: letters, digits, _, -)", http.StatusBadRequest)
			return
		}
		g := models.Group{Name: name}
		if err := db.Create(&g).Error; err != nil {
			http.Error(w, "could not create group (duplicate name?)", http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(g)
	}
}

type patchGroupBody struct {
	Name string `json:"name"`
}

// UpdateGroup renames a role and rewrites Casbin rows.
func UpdateGroup(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, err := parseUintParam(r, "groupID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body patchGroupBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if !groupNameRe.MatchString(body.Name) {
			http.Error(w, "invalid name", http.StatusBadRequest)
			return
		}
		var g models.Group
		if err := db.First(&g, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if rbac.IsSystemRole(g.Name) && body.Name != g.Name {
			http.Error(w, "cannot rename system role", http.StatusBadRequest)
			return
		}
		if body.Name == g.Name {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(g)
			return
		}
		var conflict int64
		if err := db.Model(&models.Group{}).Where("name = ? AND id <> ?", body.Name, id).Count(&conflict).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if conflict > 0 {
			http.Error(w, "name already exists", http.StatusConflict)
			return
		}
		oldName := g.Name
		g.Name = body.Name
		if err := rbac.RenameRole(e, oldName, body.Name); err != nil {
			http.Error(w, "casbin rename failed", http.StatusInternalServerError)
			return
		}
		if err := db.Save(&g).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	}
}

// DeleteGroup removes a role if unused and not a system role.
func DeleteGroup(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, err := parseUintParam(r, "groupID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var g models.Group
		if err := db.First(&g, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if rbac.IsSystemRole(g.Name) {
			http.Error(w, "cannot delete system role", http.StatusBadRequest)
			return
		}
		c := db.Model(&g).Association("Users").Count()
		if c > 0 {
			http.Error(w, "role still assigned to users", http.StatusConflict)
			return
		}
		if err := rbac.RemoveRoleFromCasbin(e, g.Name); err != nil {
			http.Error(w, "casbin error", http.StatusInternalServerError)
			return
		}
		if err := db.Delete(&g).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type policiesPayload struct {
	Policies []struct {
		Object string `json:"object"`
		Action string `json:"action"`
	} `json:"policies"`
}

// GetGroupPolicies returns Casbin p-rules for this role (by group id → name).
func GetGroupPolicies(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "groupID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var g models.Group
		if err := db.First(&g, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		rows, err := e.GetFilteredPolicy(0, g.Name)
		if err != nil {
			http.Error(w, "casbin error", http.StatusInternalServerError)
			return
		}
		type row struct {
			Object string `json:"object"`
			Action string `json:"action"`
		}
		out := make([]row, 0, len(rows))
		for _, p := range rows {
			if len(p) < 3 {
				continue
			}
			out = append(out, row{Object: p[1], Action: p[2]})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"policies": out})
	}
}

// PutGroupPolicies replaces Casbin p-rules for this role.
func PutGroupPolicies(db *gorm.DB, e *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, err := parseUintParam(r, "groupID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var g models.Group
		if err := db.First(&g, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		var body policiesPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		pairs := make([][2]string, 0, len(body.Policies))
		for _, p := range body.Policies {
			if p.Object == "" || p.Action == "" {
				continue
			}
			if p.Object != "*" && (len(p.Object) < 1 || p.Object[0] != '/') {
				http.Error(w, "object must start with / or be exactly *", http.StatusBadRequest)
				return
			}
			pairs = append(pairs, [2]string{p.Object, p.Action})
		}
		if err := rbac.SetRolePolicies(e, g.Name, pairs); err != nil {
			http.Error(w, "casbin error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}
