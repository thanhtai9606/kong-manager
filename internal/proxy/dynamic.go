package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kong/kong-manager/internal/config"
	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

// DynamicKongHandler proxies to Kong Admin: /kong-admin/... uses the "default" cluster row;
// /kong-admin/c/{slug}/... looks up KongCluster by slug.
func DynamicKongHandler(db *gorm.DB, cfg *config.Config) http.Handler {
	defaultToken := cfg.KongAdminToken
	fallbackBase, err := url.Parse(strings.TrimSpace(cfg.KongAdminURL))
	if err != nil || fallbackBase.Host == "" {
		fallbackBase, _ = url.Parse("http://127.0.0.1:8001")
	}
	prefix := strings.TrimSuffix(strings.TrimSpace(cfg.KongProxyPrefix), "/")
	if prefix == "" {
		prefix = "/kong-admin"
	}
	cPrefix := prefix + "/c/"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		var target *url.URL
		var suffix string
		token := defaultToken

		switch {
		case strings.HasPrefix(p, cPrefix):
			rest := strings.TrimPrefix(p, cPrefix)
			if rest == "" {
				http.Error(w, "cluster slug required", http.StatusBadRequest)
				return
			}
			i := strings.Index(rest, "/")
			var slug string
			if i < 0 {
				slug = rest
				suffix = "/"
			} else {
				slug = rest[:i]
				suffix = rest[i:]
			}
			var cl models.KongCluster
			if err := db.Where("slug = ? AND enabled = ?", slug, true).First(&cl).Error; err != nil {
				http.Error(w, "unknown or disabled cluster", http.StatusNotFound)
				return
			}
			u, err := url.Parse(strings.TrimRight(strings.TrimSpace(cl.AdminBaseURL), "/"))
			if err != nil || u.Host == "" {
				http.Error(w, "bad cluster url", http.StatusInternalServerError)
				return
			}
			target = u
			if strings.TrimSpace(cl.AdminToken) != "" {
				token = cl.AdminToken
			}
		case p == prefix || strings.HasPrefix(p, prefix+"/"):
			suffix = strings.TrimPrefix(p, prefix)
			if suffix == "" {
				suffix = "/"
			} else if !strings.HasPrefix(suffix, "/") {
				suffix = "/" + suffix
			}
			var cl models.KongCluster
			if err := db.Where("slug = ? AND enabled = ?", "default", true).First(&cl).Error; err != nil {
				target = fallbackBase
			} else {
				u, err := url.Parse(strings.TrimRight(strings.TrimSpace(cl.AdminBaseURL), "/"))
				if err != nil || u.Host == "" {
					target = fallbackBase
				} else {
					target = u
				}
				if strings.TrimSpace(cl.AdminToken) != "" {
					token = cl.AdminToken
				}
			}
		default:
			http.NotFound(w, r)
			return
		}

		rp := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = suffix
				req.URL.RawQuery = r.URL.RawQuery
				if token != "" {
					req.Header.Set("Kong-Admin-Token", token)
				}
				if req.Header.Get("X-Forwarded-Host") == "" {
					req.Header.Set("X-Forwarded-Host", r.Host)
				}
			},
		}
		rp.ServeHTTP(w, r)
	})
}
