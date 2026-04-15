package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/kong/kong-manager/internal/config"
	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

// KongAdminTransport builds the HTTP client used to reach Kong Admin (HTTPS).
// If KONG_UPSTREAM_TLS_SKIP_VERIFY is true, verification is disabled (dev only).
// Else if KONG_UPSTREAM_TLS_CA_FILE is set, that PEM is merged into the system root pool (production-friendly).
// Otherwise the default system roots are used.
func KongAdminTransport(cfg *config.Config) (http.RoundTripper, error) {
	if cfg.KongUpstreamTLSSkipVerify {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // gated by KONG_UPSTREAM_TLS_SKIP_VERIFY
			MinVersion:         tls.VersionTLS12,
		}
		return t, nil
	}
	caPath := strings.TrimSpace(cfg.KongUpstreamTLSCAFile)
	if caPath == "" {
		return http.DefaultTransport, nil
	}
	pemData, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("KONG_UPSTREAM_TLS_CA_FILE: %w", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("KONG_UPSTREAM_TLS_CA_FILE: no valid PEM certificates in %s", caPath)
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}
	return t, nil
}

// DynamicKongHandler proxies to Kong Admin: /kong-admin/... uses the "default" cluster row;
// /kong-admin/c/{slug}/... looks up KongCluster by slug.
func DynamicKongHandler(db *gorm.DB, cfg *config.Config, rt http.RoundTripper) http.Handler {
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

		upstream := target.String()
		rp := &httputil.ReverseProxy{
			Transport: rt,
			Director: func(req *http.Request) {
				// Incoming req still has Host from the client (e.g. localhost:8081).
				// http.Transport prefers req.Host over req.URL.Host when writing the request line;
				// without this, Kong / ingress see Host: localhost and return 404 or mis-route.
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = suffix
				req.URL.RawQuery = r.URL.RawQuery
				req.Host = target.Host
				if token != "" {
					req.Header.Set("Kong-Admin-Token", token)
				}
				if req.Header.Get("X-Forwarded-Host") == "" {
					req.Header.Set("X-Forwarded-Host", r.Host)
				}
			},
			// Default ReverseProxy returns 502 with an empty body; surface the real error for ops/debugging.
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("kong-admin proxy: upstream=%s path=%s: %v", upstream, suffix, err)
				http.Error(w, "bad gateway (upstream Kong Admin): "+err.Error(), http.StatusBadGateway)
			},
		}
		rp.ServeHTTP(w, r)
	})
}
