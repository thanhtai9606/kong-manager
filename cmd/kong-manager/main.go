package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kong/kong-manager/internal/auth"
	"github.com/kong/kong-manager/internal/bootstrap"
	"github.com/kong/kong-manager/internal/config"
	appdb "github.com/kong/kong-manager/internal/db"
	"github.com/kong/kong-manager/internal/httpapi"
	"github.com/kong/kong-manager/internal/proxy"
	"github.com/kong/kong-manager/internal/rbac"
)

func main() {
	cfg := config.Load()

	db, err := appdb.Open(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := appdb.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	enforcer, err := rbac.NewEnforcer(db)
	if err != nil {
		log.Fatalf("casbin: %v", err)
	}
	if err := bootstrap.Run(cfg, db, enforcer); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	kongURL, err := url.Parse(cfg.KongAdminURL)
	if err != nil {
		log.Fatalf("KONG_ADMIN_URL: %v", err)
	}

	jwtSvc := auth.NewService(cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	r.Post("/api/auth/login", jwtSvc.LoginHandler(db))

	kongHandler := httpapi.JWTAuth(jwtSvc)(
		httpapi.CasbinAuthorize(enforcer)(
			proxy.Handler(kongURL, cfg.KongAdminToken, cfg.KongProxyPrefix),
		),
	)

	prefix := strings.TrimSuffix(cfg.KongProxyPrefix, "/")
	if prefix == "" {
		prefix = "/kong-admin"
	}

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			kongHandler.ServeHTTP(w, r)
			return
		}
		httpapi.SPA(cfg.StaticDir).ServeHTTP(w, r)
	}))

	log.Printf("listening on %s (static=%s kong=%s)", cfg.HTTPAddr, cfg.StaticDir, cfg.KongAdminURL)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatal(err)
	}
}
