package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/kong/kong-manager/internal/models"
	"gorm.io/gorm"
)

// MeHandler returns the authenticated user and their groups (Casbin roles).
func MeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFrom(r.Context())
		if !ok || claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var user models.User
		if err := db.Preload("Groups").First(&user, claims.UserID).Error; err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		type groupOut struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}
		groups := make([]groupOut, 0, len(user.Groups))
		for _, g := range user.Groups {
			groups = append(groups, groupOut{ID: g.ID, Name: g.Name})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"groups":   groups,
		})
	}
}
