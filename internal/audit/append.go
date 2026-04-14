package audit

import (
	"net"
	"net/http"
	"strings"

	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

// ClientIP returns a best-effort client address for logging.
func ClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Append persists one audit row. Failures are returned to the caller.
func Append(db *gorm.DB, r *http.Request, actorUsername, action, resource, details string) error {
	row := models.AuditLog{
		ActorUsername: actorUsername,
		Action:        action,
		Resource:      resource,
		Details:       details,
		ClientIP:      ClientIP(r),
	}
	return db.Create(&row).Error
}
