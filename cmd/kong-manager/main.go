package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kong/kong-manager/internal/admin"
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

	jwtSvc := auth.NewService(cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Align request path with Casbin policies when the SPA uses a non-root asset base (ADMIN_GUI_PATH / kconfig).
	r.Use(httpapi.StripGUIPath(cfg.AdminGUIPath))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	r.Post("/api/auth/login", jwtSvc.LoginHandler(db))
	r.Get("/api/auth/sso/providers", jwtSvc.PublicSSOProvidersHandler(db))
	r.Get("/api/auth/oidc/{slug}/login", jwtSvc.OIDCLoginHandler(db))
	r.Get("/api/auth/oidc/{slug}/callback", jwtSvc.OIDCCallbackHandler(db, enforcer))

	r.Route("/api/admin", func(ar chi.Router) {
		ar.Use(httpapi.JWTAuth(jwtSvc))
		ar.Use(httpapi.CasbinAuthorize(enforcer, cfg.KongProxyPrefix))
		ar.Put("/users/{userID}/groups", admin.PutUserGroups(db, enforcer))
		ar.Patch("/users/{userID}", admin.PatchUser(db, enforcer))
		ar.Delete("/users/{userID}", admin.DeleteUser(db, enforcer))
		ar.Post("/users", admin.CreateUser(db, enforcer))
		ar.Get("/users", admin.ListUsers(db))
		ar.Get("/audit-logs", admin.ListAuditLogs(db))
		ar.Get("/groups", admin.ListGroups(db))
		ar.Post("/groups", admin.CreateGroup(db))
		ar.Get("/groups/{groupID}/policies", admin.GetGroupPolicies(db, enforcer))
		ar.Put("/groups/{groupID}/policies", admin.PutGroupPolicies(db, enforcer))
		ar.Patch("/groups/{groupID}", admin.UpdateGroup(db, enforcer))
		ar.Delete("/groups/{groupID}", admin.DeleteGroup(db, enforcer))
		ar.Get("/rbac", admin.RBACSnapshot(enforcer))
		ar.Get("/kong-clusters", admin.ListKongClusters(db))
		ar.Post("/kong-clusters", admin.CreateKongCluster(db))
		ar.Patch("/kong-clusters/{clusterID}", admin.PatchKongCluster(db))
		ar.Delete("/kong-clusters/{clusterID}", admin.DeleteKongCluster(db))
		ar.Get("/sso-providers", admin.ListSSOProviders(db))
		ar.Post("/sso-providers", admin.CreateSSOProvider(db))
		ar.Patch("/sso-providers/{ssoProviderID}", admin.PatchSSOProvider(db))
		ar.Delete("/sso-providers/{ssoProviderID}", admin.DeleteSSOProvider(db))
	})

	kongHandler := httpapi.JWTAuth(jwtSvc)(
		httpapi.CasbinAuthorize(enforcer, cfg.KongProxyPrefix)(
			proxy.DynamicKongHandler(db, cfg),
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

	log.Printf("listening on %s (static=%s default_kong=%s)", cfg.HTTPAddr, cfg.StaticDir, cfg.KongAdminURL)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatal(err)
	}
}
