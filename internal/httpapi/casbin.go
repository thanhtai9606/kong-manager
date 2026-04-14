package httpapi

import (
	"net/http"

	"github.com/casbin/casbin/v2"
)

// CasbinAuthorize enforces RBAC for the Kong proxy path.
func CasbinAuthorize(e *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFrom(r.Context())
			if !ok || claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			allowed, err := e.Enforce(claims.Username, r.URL.Path, r.Method)
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
