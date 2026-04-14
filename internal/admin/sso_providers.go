package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

type ssoProviderDTO struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	IssuerURL string `json:"issuer_url"`
	ClientID  string `json:"client_id"`
	HasSecret bool   `json:"has_secret"`
	Scopes    string `json:"scopes"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

func toSSODTO(p models.SSOProvider) ssoProviderDTO {
	return ssoProviderDTO{
		ID:        p.ID,
		Slug:      p.Slug,
		Name:      p.Name,
		IssuerURL: p.IssuerURL,
		ClientID:  p.ClientID,
		HasSecret: strings.TrimSpace(p.ClientSecret) != "",
		Scopes:    p.Scopes,
		Enabled:   p.Enabled,
		SortOrder: p.SortOrder,
	}
}

// ListSSOProviders returns all SSO provider rows (secrets never exposed).
func ListSSOProviders(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rows []models.SSOProvider
		if err := db.Order("sort_order, name").Find(&rows).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out := make([]ssoProviderDTO, 0, len(rows))
		for _, p := range rows {
			out = append(out, toSSODTO(p))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

type createSSOProviderBody struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	IssuerURL    string `json:"issuer_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scopes       string `json:"scopes"`
	Enabled      *bool  `json:"enabled"`
	SortOrder    int    `json:"sort_order"`
}

// CreateSSOProvider adds an OIDC provider.
func CreateSSOProvider(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body createSSOProviderBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		body.Slug = strings.TrimSpace(body.Slug)
		body.Name = strings.TrimSpace(body.Name)
		body.IssuerURL = strings.TrimSpace(body.IssuerURL)
		body.ClientID = strings.TrimSpace(body.ClientID)
		body.ClientSecret = strings.TrimSpace(body.ClientSecret)
		if body.Name == "" || body.Slug == "" || body.IssuerURL == "" || body.ClientID == "" || body.ClientSecret == "" {
			http.Error(w, "name, slug, issuer_url, client_id, client_secret required", http.StatusBadRequest)
			return
		}
		if !slugRe.MatchString(body.Slug) {
			http.Error(w, "invalid slug", http.StatusBadRequest)
			return
		}
		en := true
		if body.Enabled != nil {
			en = *body.Enabled
		}
		p := models.SSOProvider{
			Slug:         body.Slug,
			Name:         body.Name,
			IssuerURL:    strings.TrimRight(body.IssuerURL, "/"),
			ClientID:     body.ClientID,
			ClientSecret: body.ClientSecret,
			Scopes:       strings.TrimSpace(body.Scopes),
			Enabled:      en,
			SortOrder:    body.SortOrder,
		}
		if err := db.Create(&p).Error; err != nil {
			http.Error(w, "could not create (duplicate slug?)", http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toSSODTO(p))
	}
}

type patchSSOProviderBody struct {
	Name         *string `json:"name"`
	IssuerURL    *string `json:"issuer_url"`
	ClientID     *string `json:"client_id"`
	ClientSecret *string `json:"client_secret"`
	Scopes       *string `json:"scopes"`
	Enabled      *bool   `json:"enabled"`
	SortOrder    *int    `json:"sort_order"`
	ClearSecret  *bool   `json:"clear_secret"`
}

// PatchSSOProvider updates a provider.
func PatchSSOProvider(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "ssoProviderID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body patchSSOProviderBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var p models.SSOProvider
		if err := db.First(&p, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		updates := map[string]any{}
		if body.Name != nil {
			s := strings.TrimSpace(*body.Name)
			if s == "" {
				http.Error(w, "name empty", http.StatusBadRequest)
				return
			}
			updates["name"] = s
		}
		if body.IssuerURL != nil {
			s := strings.TrimSpace(*body.IssuerURL)
			if s == "" {
				http.Error(w, "issuer_url empty", http.StatusBadRequest)
				return
			}
			updates["issuer_url"] = strings.TrimRight(s, "/")
		}
		if body.ClientID != nil {
			s := strings.TrimSpace(*body.ClientID)
			if s == "" {
				http.Error(w, "client_id empty", http.StatusBadRequest)
				return
			}
			updates["client_id"] = s
		}
		if body.ClientSecret != nil {
			s := strings.TrimSpace(*body.ClientSecret)
			if s != "" {
				updates["client_secret"] = s
			}
		}
		if body.ClearSecret != nil && *body.ClearSecret {
			updates["client_secret"] = ""
		}
		if body.Scopes != nil {
			updates["scopes"] = strings.TrimSpace(*body.Scopes)
		}
		if body.Enabled != nil {
			updates["enabled"] = *body.Enabled
		}
		if body.SortOrder != nil {
			updates["sort_order"] = *body.SortOrder
		}
		if len(updates) == 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(toSSODTO(p))
			return
		}
		if err := db.Model(&p).Updates(updates).Error; err != nil {
			http.Error(w, "update error", http.StatusInternalServerError)
			return
		}
		if err := db.First(&p, id).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toSSODTO(p))
	}
}

// DeleteSSOProvider soft-deletes a provider.
func DeleteSSOProvider(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "ssoProviderID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		res := db.Delete(&models.SSOProvider{}, id)
		if res.Error != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if res.RowsAffected == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
