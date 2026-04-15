package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration for the API server.
type Config struct {
	HTTPAddr           string
	StaticDir          string
	DatabaseDriver     string
	DatabaseURL        string
	JWTSecret          string
	JWTTTL time.Duration
	// KongAdminURL seeds the initial "default" KongCluster row and is a fallback if that row is missing.
	// At runtime, each cluster’s upstream is kong_clusters.admin_base_url (see proxy.DynamicKongHandler).
	KongAdminURL string
	// KongAdminToken is the default Admin token when a cluster row has no per-cluster token.
	KongAdminToken string
	KongProxyPrefix    string
	// KongUpstreamTLSSkipVerify disables TLS verification when the BFF connects to Kong Admin (HTTPS).
	// Set via KONG_UPSTREAM_TLS_SKIP_VERIFY=true for local/dev Kong on https://localhost:8444 with self-signed certs.
	// Do not enable in production unless you trust the network path to Kong Admin.
	KongUpstreamTLSSkipVerify bool
	// AdminGUIPath is the SPA base path (e.g. /__km_base__) without trailing slash; empty if served at /.
	// Must match window.K_CONFIG.ADMIN_GUI_PATH so API paths are normalized before Casbin.
	AdminGUIPath       string
	BootstrapUsername  string
	BootstrapPassword  string
	CookieSecure       bool
	// PublicBaseURL is the browser-facing origin (scheme + host, no path), e.g. https://kong-manager.example.com
	// Used to build OIDC redirect_uri when the request Host does not match the browser (e.g. some dev proxies
	// rewrite Host to the upstream port). Set PUBLIC_BASE_URL=http://localhost:8080 when FE is on :8080 and
	// BFF on :8081 if redirect_uri still does not match Keycloak.
	PublicBaseURL string
	// OIDCTLSkipVerify disables TLS certificate verification for outbound OIDC discovery/token/JWKS (Issuer URL).
	// Set OIDC_TLS_SKIP_VERIFY=true for local IdPs with self-signed certs (e.g. Keycloak on https://localhost).
	// Do not use in production.
	OIDCTLSkipVerify bool
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustParseDuration(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

func parseBool(s string, def bool) bool {
	if s == "" {
		return def
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}

// adminGUIPath returns the SPA asset prefix (no trailing slash) for StripGUIPath.
// Must match Vite `base` in production (default /__km_base__ in vite.config.ts).
// If ADMIN_GUI_PATH is unset, use that default; if set to empty string explicitly, use "" (root build, e.g. DISABLE_BASE_PATH=true).
func adminGUIPath() string {
	v, ok := os.LookupEnv("ADMIN_GUI_PATH")
	if !ok {
		return "/__km_base__"
	}
	return strings.TrimSpace(v)
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		StaticDir:       getenv("STATIC_DIR", "./dist"),
		DatabaseDriver:  getenv("DATABASE_DRIVER", "sqlite"),
		DatabaseURL:     getenv("DATABASE_URL", "file:kong-manager.db?cache=shared&mode=rwc"),
		JWTSecret:       getenv("JWT_SECRET", "dev-insecure-change-me"),
		JWTTTL:          mustParseDuration(getenv("JWT_TTL", "24h"), 24*time.Hour),
		KongAdminURL:    getenv("KONG_ADMIN_URL", "http://127.0.0.1:8001"),
		KongAdminToken:  getenv("KONG_ADMIN_TOKEN", ""),
		KongProxyPrefix: getenv("KONG_PROXY_PREFIX", "/kong-admin"),
		KongUpstreamTLSSkipVerify: parseBool(getenv("KONG_UPSTREAM_TLS_SKIP_VERIFY", ""), false),
		AdminGUIPath:    adminGUIPath(),
		BootstrapUsername: getenv("BOOTSTRAP_ADMIN_USERNAME", ""),
		BootstrapPassword: getenv("BOOTSTRAP_ADMIN_PASSWORD", ""),
		CookieSecure:      parseBool(getenv("COOKIE_SECURE", "false"), false),
		PublicBaseURL:     strings.TrimSpace(getenv("PUBLIC_BASE_URL", "")),
		OIDCTLSkipVerify:  parseBool(getenv("OIDC_TLS_SKIP_VERIFY", ""), false),
	}
}
