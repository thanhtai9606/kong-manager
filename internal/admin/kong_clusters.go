package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

type kongClusterDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	AdminBaseURL string `json:"admin_base_url"`
	HasToken     bool   `json:"has_token"`
	Enabled      bool   `json:"enabled"`
	SortOrder    int    `json:"sort_order"`
}

func toDTO(c models.KongCluster) kongClusterDTO {
	return kongClusterDTO{
		ID:           c.ID,
		Name:         c.Name,
		Slug:         c.Slug,
		AdminBaseURL: c.AdminBaseURL,
		HasToken:     strings.TrimSpace(c.AdminToken) != "",
		Enabled:      c.Enabled,
		SortOrder:    c.SortOrder,
	}
}

// ListKongClusters returns all Kong clusters (no secrets).
func ListKongClusters(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rows []models.KongCluster
		if err := db.Order("sort_order, name").Find(&rows).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out := make([]kongClusterDTO, 0, len(rows))
		for _, c := range rows {
			out = append(out, toDTO(c))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

type createKongClusterBody struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	AdminBaseURL string `json:"admin_base_url"`
	AdminToken   string `json:"admin_token"`
	SortOrder    int    `json:"sort_order"`
}

// CreateKongCluster adds a cluster row.
func CreateKongCluster(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body createKongClusterBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Slug = strings.TrimSpace(body.Slug)
		body.AdminBaseURL = strings.TrimSpace(body.AdminBaseURL)
		if body.Name == "" || body.Slug == "" || body.AdminBaseURL == "" {
			http.Error(w, "name, slug, admin_base_url required", http.StatusBadRequest)
			return
		}
		if !slugRe.MatchString(body.Slug) {
			http.Error(w, "invalid slug", http.StatusBadRequest)
			return
		}
		c := models.KongCluster{
			Name:         body.Name,
			Slug:         body.Slug,
			AdminBaseURL: strings.TrimRight(body.AdminBaseURL, "/"),
			AdminToken:   body.AdminToken,
			Enabled:      true,
			SortOrder:    body.SortOrder,
		}
		if err := db.Create(&c).Error; err != nil {
			http.Error(w, "could not create (duplicate slug?)", http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toDTO(c))
	}
}

type patchKongClusterBody struct {
	Name         *string `json:"name"`
	AdminBaseURL *string `json:"admin_base_url"`
	AdminToken   *string `json:"admin_token"`
	Enabled      *bool   `json:"enabled"`
	SortOrder    *int    `json:"sort_order"`
	ClearToken   *bool   `json:"clear_token"`
}

// PatchKongCluster updates a cluster.
func PatchKongCluster(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "clusterID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body patchKongClusterBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var c models.KongCluster
		if err := db.First(&c, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if body.Name != nil {
			c.Name = strings.TrimSpace(*body.Name)
		}
		if body.AdminBaseURL != nil {
			c.AdminBaseURL = strings.TrimRight(strings.TrimSpace(*body.AdminBaseURL), "/")
		}
		if body.Enabled != nil {
			c.Enabled = *body.Enabled
		}
		if body.SortOrder != nil {
			c.SortOrder = *body.SortOrder
		}
		if body.ClearToken != nil && *body.ClearToken {
			c.AdminToken = ""
		} else if body.AdminToken != nil {
			c.AdminToken = *body.AdminToken
		}
		if err := db.Save(&c).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toDTO(c))
	}
}

// DeleteKongCluster removes a cluster (not the default slug).
func DeleteKongCluster(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "clusterID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var c models.KongCluster
		if err := db.First(&c, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if c.Slug == "default" {
			http.Error(w, "cannot delete the default cluster", http.StatusBadRequest)
			return
		}
		var total int64
		if err := db.Model(&models.KongCluster{}).Count(&total).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if total <= 1 {
			http.Error(w, "cannot delete the last cluster", http.StatusBadRequest)
			return
		}
		if err := db.Delete(&c).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
