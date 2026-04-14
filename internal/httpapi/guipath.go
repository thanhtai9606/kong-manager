package httpapi

import (
	"net/http"
	"strings"
)

// StripGUIPath rewrites r.URL.Path by removing a static GUI prefix (e.g. /__km_base__)
// so Casbin policies (/kong-admin/*, /api/admin/*) and chi routes match the same paths
// the SPA would use without the asset base segment.
func StripGUIPath(guiPrefix string) func(http.Handler) http.Handler {
	p := strings.TrimSpace(guiPrefix)
	p = strings.TrimSuffix(p, "/")
	if p == "" || p == "/" {
		return func(next http.Handler) http.Handler { return next }
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, p) {
				rest := strings.TrimPrefix(r.URL.Path, p)
				if rest == "" {
					rest = "/"
				} else if !strings.HasPrefix(rest, "/") {
					rest = "/" + rest
				}
				r.URL.Path = rest
			}
			next.ServeHTTP(w, r)
		})
	}
}
