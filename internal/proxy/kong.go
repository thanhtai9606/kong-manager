package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Handler forwards requests to Kong Admin API after stripping the proxy prefix.
func Handler(kongBase *url.URL, kongAdminToken string, proxyPrefix string) http.Handler {
	p := strings.TrimSuffix(proxyPrefix, "/")
	if p == "" {
		p = "/kong-admin"
	}

	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			clientHost := req.Host
			req.URL.Scheme = kongBase.Scheme
			req.URL.Host = kongBase.Host
			suffix := strings.TrimPrefix(req.URL.Path, p)
			if suffix == "" {
				suffix = "/"
			}
			if !strings.HasPrefix(suffix, "/") {
				suffix = "/" + suffix
			}
			req.URL.Path = suffix
			req.Host = kongBase.Host
			if kongAdminToken != "" {
				req.Header.Set("Kong-Admin-Token", kongAdminToken)
			}
			if req.Header.Get("X-Forwarded-Host") == "" {
				req.Header.Set("X-Forwarded-Host", clientHost)
			}
		},
	}
	return rp
}
