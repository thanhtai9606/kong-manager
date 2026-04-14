package httpapi

import "strings"

// KongPathForPolicy maps a proxied path to the legacy shape used in Casbin policies.
// /kong-admin/c/{slug}/services -> /kong-admin/services so existing /kong-admin/* rules still apply.
func KongPathForPolicy(reqPath string, kongPrefix string) string {
	p := strings.TrimSuffix(strings.TrimSpace(kongPrefix), "/")
	if p == "" {
		p = "/kong-admin"
	}
	cPrefix := p + "/c/"
	if !strings.HasPrefix(reqPath, cPrefix) {
		return reqPath
	}
	rest := strings.TrimPrefix(reqPath, cPrefix)
	if rest == "" {
		return p
	}
	i := strings.Index(rest, "/")
	if i < 0 {
		return p
	}
	return p + rest[i:]
}
