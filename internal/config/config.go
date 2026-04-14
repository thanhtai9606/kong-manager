package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration for the API server.
type Config struct {
	HTTPAddr           string
	StaticDir          string
	DatabaseDriver     string
	DatabaseURL        string
	JWTSecret          string
	JWTTTL             time.Duration
	KongAdminURL       string
	KongAdminToken     string
	KongProxyPrefix    string
	BootstrapUsername  string
	BootstrapPassword  string
	CookieSecure       bool
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
		BootstrapUsername: getenv("BOOTSTRAP_ADMIN_USERNAME", ""),
		BootstrapPassword: getenv("BOOTSTRAP_ADMIN_PASSWORD", ""),
		CookieSecure:      parseBool(getenv("COOKIE_SECURE", "false"), false),
	}
}
