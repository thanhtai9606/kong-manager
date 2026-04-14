package httpapi

import (
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
)

// CasbinAuthorize enforces RBAC for the Kong proxy path.
// kongProxyPrefix should match config (e.g. /kong-admin) so /kong-admin/c/{slug}/... normalizes for policy checks.
func CasbinAuthorize(e *casbin.Enforcer, kongProxyPrefix string) func(http.Handler) http.Handler {
	kongProxyPrefix = strings.TrimSpace(kongProxyPrefix)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFrom(r.Context())
			if !ok || claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			obj := KongPathForPolicy(r.URL.Path, kongProxyPrefix)
			allowed, err := e.Enforce(claims.Username, obj, r.Method)
			if err != nil {
				http.Error(w, "authorization error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
