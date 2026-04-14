package httpapi

import "testing"

func TestKongPathForPolicy(t *testing.T) {
	const p = "/kong-admin"
	if got := KongPathForPolicy("/kong-admin/services", p); got != "/kong-admin/services" {
		t.Fatalf("legacy: got %q", got)
	}
	if got := KongPathForPolicy("/kong-admin/c/prod/services", p); got != "/kong-admin/services" {
		t.Fatalf("cluster: got %q want /kong-admin/services", got)
	}
	if got := KongPathForPolicy("/kong-admin/c/prod", p); got != "/kong-admin" {
		t.Fatalf("cluster root: got %q", got)
	}
}
