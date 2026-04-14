package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

const (
	defaultAuditLimit  = 50
	maxAuditLimit      = 200
	defaultAuditOffset = 0
)

type auditLogListResponse struct {
	Items []models.AuditLog `json:"items"`
	Total int64             `json:"total"`
}

// ListAuditLogs returns recent audit rows (newest first).
func ListAuditLogs(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		limit := parseIntQuery(r, "limit", defaultAuditLimit)
		if limit < 1 {
			limit = defaultAuditLimit
		}
		if limit > maxAuditLimit {
			limit = maxAuditLimit
		}
		offset := parseIntQuery(r, "offset", defaultAuditOffset)
		if offset < 0 {
			offset = 0
		}
		var total int64
		if err := db.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		var items []models.AuditLog
		if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auditLogListResponse{Items: items, Total: total})
	}
}

func parseIntQuery(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
