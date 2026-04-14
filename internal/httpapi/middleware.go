package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/kong/kong-manager/internal/auth"
)

type ctxKey int

const claimsKey ctxKey = 1

// WithClaims attaches JWT claims to the request context.
func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// ClaimsFrom returns claims set by JWT middleware.
func ClaimsFrom(ctx context.Context) (*auth.Claims, bool) {
	v := ctx.Value(claimsKey)
	if v == nil {
		return nil, false
	}
	c, ok := v.(*auth.Claims)
	return c, ok
}

// JWTAuth validates Authorization: Bearer <token>.
func JWTAuth(s *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			const pfx = "bearer "
			if len(h) < len(pfx) || !strings.EqualFold(h[:len(pfx)], pfx) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimSpace(h[len(pfx):])
			claims, err := s.Parse(raw)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
