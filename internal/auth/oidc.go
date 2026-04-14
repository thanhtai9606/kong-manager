package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/kong/kong-manager/internal/config"
	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/rbac"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const oidcNonceCookie = "km_oidc_nonce"

// PublicSSOProviders returns enabled IdPs for the login screen (no secrets).
func (s *Service) PublicSSOProvidersHandler(db *gorm.DB) http.HandlerFunc {
	type row struct {
		Slug      string `json:"slug"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var providers []models.SSOProvider
		if err := db.Where("enabled = ?", true).Order("sort_order, name").Find(&providers).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out := make([]row, 0, len(providers))
		for _, p := range providers {
			out = append(out, row{Slug: p.Slug, Name: p.Name, SortOrder: p.SortOrder})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func publicOrigin(r *http.Request, cfg *config.Config) string {
	if u := strings.TrimSpace(cfg.PublicBaseURL); u != "" {
		return strings.TrimRight(u, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == "https" {
		scheme = "https"
	} else if p == "http" {
		scheme = "http"
	}
	host := r.Host
	if xh := r.Header.Get("X-Forwarded-Host"); xh != "" {
		host = strings.TrimSpace(strings.Split(xh, ",")[0])
	}
	return scheme + "://" + host
}

func oidcAppPrefix(cfg *config.Config) string {
	p := strings.TrimSpace(cfg.AdminGUIPath)
	p = strings.TrimSuffix(p, "/")
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// buildOIDCRedirectURI is the exact browser URL the IdP must allow (includes ADMIN_GUI_PATH when set).
func buildOIDCRedirectURI(cfg *config.Config, r *http.Request, slug string) string {
	base := publicOrigin(r, cfg) + oidcAppPrefix(cfg)
	return base + "/api/auth/oidc/" + url.PathEscape(slug) + "/callback"
}

func loginPageBase(cfg *config.Config, r *http.Request) string {
	origin := publicOrigin(r, cfg)
	gui := strings.TrimRight(cfg.AdminGUIPath, "/")
	if gui != "" {
		return origin + gui
	}
	return origin
}

func redirectToLogin(cfg *config.Config, r *http.Request, w http.ResponseWriter, fragment string) {
	base := loginPageBase(cfg, r)
	loc := base + "/login"
	if fragment != "" {
		loc += "#" + fragment
	}
	http.Redirect(w, r, loc, http.StatusFound)
}

func redirectToLoginQuery(cfg *config.Config, r *http.Request, w http.ResponseWriter, q string) {
	base := loginPageBase(cfg, r)
	loc := base + "/login"
	if q != "" {
		loc += "?" + q
	}
	http.Redirect(w, r, loc, http.StatusFound)
}

func defaultScopes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{oidc.ScopeOpenID, "profile", "email"}
	}
	var out []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ','
	}) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{oidc.ScopeOpenID, "profile", "email"}
	}
	return out
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func randomPasswordHash() (string, error) {
	b, err := randomBytes(32)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func shortSubHash(sub string) string {
	h := sha256.Sum256([]byte(sub))
	return hex.EncodeToString(h[:8])
}

type idTokenProfile struct {
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
}

func pickUsername(p *idTokenProfile, sub string) string {
	if p.Email != "" {
		return strings.TrimSpace(p.Email)
	}
	if p.PreferredUsername != "" {
		return strings.TrimSpace(p.PreferredUsername)
	}
	return "oidc-" + shortSubHash(sub)
}

func ensureUniqueUsername(db *gorm.DB, base string) (string, error) {
	if len(base) > 160 {
		base = base[:160]
	}
	candidate := base
	for i := 0; i < 50; i++ {
		var n int64
		if err := db.Model(&models.User{}).Where("username = ?", candidate).Count(&n).Error; err != nil {
			return "", err
		}
		if n == 0 {
			return candidate, nil
		}
		suffix, err := randomBytes(3)
		if err != nil {
			return "", err
		}
		candidate = base + "-" + hex.EncodeToString(suffix)
	}
	return "", errors.New("could not allocate username")
}

// OIDCLoginHandler redirects the browser to the IdP authorize endpoint.
func (s *Service) OIDCLoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		slug := strings.TrimSpace(chi.URLParam(r, "slug"))
		if slug == "" {
			http.Error(w, "missing slug", http.StatusBadRequest)
			return
		}
		var provider models.SSOProvider
		if err := db.Where("slug = ? AND enabled = ?", slug, true).First(&provider).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "unknown provider", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		oidcProvider, err := oidc.NewProvider(ctx, strings.TrimRight(provider.IssuerURL, "/"))
		if err != nil {
			log.Printf("oidc: provider discovery %q: %v", provider.IssuerURL, err)
			http.Error(w, "oidc discovery failed", http.StatusBadGateway)
			return
		}

		nonceBytes, err := randomBytes(24)
		if err != nil {
			http.Error(w, "random error", http.StatusInternalServerError)
			return
		}
		nonce := hex.EncodeToString(nonceBytes)

		redirectURI := buildOIDCRedirectURI(s.cfg, r, slug)
		oauthCfg := oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			RedirectURL:  redirectURI,
			Endpoint:     oidcProvider.Endpoint(),
			Scopes:       defaultScopes(provider.Scopes),
		}

		state, err := s.SignOIDCState(slug, provider.ID, nonce)
		if err != nil {
			http.Error(w, "state error", http.StatusInternalServerError)
			return
		}

		authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("nonce", nonce))

		http.SetCookie(w, &http.Cookie{
			Name:     oidcNonceCookie,
			Value:    nonce,
			Path:     "/",
			MaxAge:   900,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   s.cfg.CookieSecure,
		})

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// OIDCCallbackHandler exchanges the code, verifies the ID token, and issues an app JWT.
func (s *Service) OIDCCallbackHandler(db *gorm.DB, enforcer *casbin.Enforcer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		slug := strings.TrimSpace(chi.URLParam(r, "slug"))
		if errQ := strings.TrimSpace(r.URL.Query().Get("error")); errQ != "" {
			desc := r.URL.Query().Get("error_description")
			log.Printf("oidc callback error from idp: %s %s", errQ, desc)
			redirectToLoginQuery(s.cfg, r, w, "sso_error="+url.QueryEscape(errQ))
			return
		}
		code := strings.TrimSpace(r.URL.Query().Get("code"))
		state := strings.TrimSpace(r.URL.Query().Get("state"))
		if code == "" || state == "" {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=missing_code")
			return
		}

		st, err := s.ParseOIDCState(state)
		if err != nil {
			log.Printf("oidc: bad state: %v", err)
			redirectToLoginQuery(s.cfg, r, w, "sso_error=bad_state")
			return
		}
		if st.Slug != slug {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=state_mismatch")
			return
		}

		var provider models.SSOProvider
		if err := db.First(&provider, st.Pid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				redirectToLoginQuery(s.cfg, r, w, "sso_error=unknown_provider")
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if !provider.Enabled || provider.Slug != slug {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=provider_disabled")
			return
		}

		nonceCookie, err := r.Cookie(oidcNonceCookie)
		if err != nil || nonceCookie.Value == "" {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=missing_nonce")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: oidcNonceCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})

		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()

		oidcProvider, err := oidc.NewProvider(ctx, strings.TrimRight(provider.IssuerURL, "/"))
		if err != nil {
			log.Printf("oidc: discovery: %v", err)
			redirectToLoginQuery(s.cfg, r, w, "sso_error=discovery")
			return
		}

		redirectURI := buildOIDCRedirectURI(s.cfg, r, slug)
		oauthCfg := oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			RedirectURL:  redirectURI,
			Endpoint:     oidcProvider.Endpoint(),
			Scopes:       defaultScopes(provider.Scopes),
		}

		oauth2Token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			log.Printf("oidc: token exchange: %v", err)
			redirectToLoginQuery(s.cfg, r, w, "sso_error=exchange")
			return
		}

		rawIDToken, _ := oauth2Token.Extra("id_token").(string)
		if rawIDToken == "" {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=no_id_token")
			return
		}

		verifier := oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			log.Printf("oidc: id token verify: %v", err)
			redirectToLoginQuery(s.cfg, r, w, "sso_error=id_token")
			return
		}

		if idToken.Nonce != nonceCookie.Value {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=nonce")
			return
		}

		var prof idTokenProfile
		if err := idToken.Claims(&prof); err != nil {
			log.Printf("oidc: claims: %v", err)
			redirectToLoginQuery(s.cfg, r, w, "sso_error=claims")
			return
		}

		sub := idToken.Subject
		if sub == "" {
			redirectToLoginQuery(s.cfg, r, w, "sso_error=no_sub")
			return
		}

		baseName := pickUsername(&prof, sub)
		if baseName == "" {
			baseName = "oidc-" + shortSubHash(sub)
		}

		var user models.User
		err = db.Preload("Groups").Where("sso_provider_id = ? AND external_sub = ?", provider.ID, sub).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uname, err := ensureUniqueUsername(db, baseName)
			if err != nil {
				log.Printf("oidc: username: %v", err)
				redirectToLoginQuery(s.cfg, r, w, "sso_error=username")
				return
			}
			pwHash, err := randomPasswordHash()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			pid := provider.ID
			email := strings.TrimSpace(prof.Email)
			user = models.User{
				Username:      uname,
				PasswordHash:  pwHash,
				Email:         email,
				SSOProviderID: &pid,
				ExternalSub:   sub,
			}
			if err := db.Create(&user).Error; err != nil {
				log.Printf("oidc: create user: %v", err)
				redirectToLoginQuery(s.cfg, r, w, "sso_error=create_user")
				return
			}
			var viewer models.Group
			if err := db.Where("name = ?", "viewer").First(&viewer).Error; err != nil {
				log.Printf("oidc: viewer group: %v", err)
				redirectToLoginQuery(s.cfg, r, w, "sso_error=groups")
				return
			}
			if err := db.Model(&user).Association("Groups").Append(&viewer); err != nil {
				log.Printf("oidc: append group: %v", err)
				redirectToLoginQuery(s.cfg, r, w, "sso_error=groups")
				return
			}
			if err := rbac.SyncUserGroups(enforcer, user.Username, []string{"viewer"}); err != nil {
				log.Printf("oidc: casbin: %v", err)
				redirectToLoginQuery(s.cfg, r, w, "sso_error=rbac")
				return
			}
		} else if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		} else {
			changed := false
			if em := strings.TrimSpace(prof.Email); em != "" && user.Email != em {
				user.Email = em
				changed = true
			}
			if changed {
				_ = db.Model(&user).Updates(map[string]any{"email": user.Email}).Error
			}
		}

		tokens, err := s.IssueAccessToken(user.ID, user.Username)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		frag := "km_token=" + url.QueryEscape(tokens.AccessToken)
		redirectToLogin(s.cfg, r, w, frag)
	}
}
