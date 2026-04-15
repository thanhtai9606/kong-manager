package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kong/kong-manager/internal/models"
	"github.com/kong/kong-manager/internal/notify"
	"gorm.io/gorm"
)

var notificationTypes = map[string]struct{}{
	"slack": {}, "teams": {}, "telegram": {}, "email": {},
}

func validNotificationType(t string) bool {
	_, ok := notificationTypes[strings.ToLower(strings.TrimSpace(t))]
	return ok
}

func normalizeConfigJSON(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "{}", nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return "", err
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func configToRawMessage(s string) json.RawMessage {
	s = strings.TrimSpace(s)
	if s == "" {
		return json.RawMessage("{}")
	}
	if !json.Valid([]byte(s)) {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}

type notificationChannelDTO struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Config    json.RawMessage `json:"config"`
	HasSecret bool            `json:"has_secret"`
	Enabled   bool            `json:"enabled"`
	SortOrder int             `json:"sort_order"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func toNotificationDTO(ch models.NotificationChannel) notificationChannelDTO {
	return notificationChannelDTO{
		ID:        ch.ID,
		Name:      ch.Name,
		Type:      ch.Type,
		Config:    configToRawMessage(ch.ConfigJSON),
		HasSecret: strings.TrimSpace(ch.Secret) != "",
		Enabled:   ch.Enabled,
		SortOrder: ch.SortOrder,
		CreatedAt: ch.CreatedAt.UTC(),
		UpdatedAt: ch.UpdatedAt.UTC(),
	}
}

func validateMergedState(typ, cfgJSON, secret string) error {
	typ = strings.ToLower(strings.TrimSpace(typ))
	secret = strings.TrimSpace(secret)
	cfgJSON = strings.TrimSpace(cfgJSON)
	if cfgJSON == "" {
		cfgJSON = "{}"
	}

	switch typ {
	case "slack", "teams", "telegram", "email":
		if secret == "" {
			return errors.New("secret is required for this channel type")
		}
	}

	switch typ {
	case "telegram":
		var m map[string]any
		if err := json.Unmarshal([]byte(cfgJSON), &m); err != nil {
			return errors.New("invalid telegram config")
		}
		if strings.TrimSpace(chatIDFromConfig(m)) == "" {
			return errors.New("telegram config.chat_id is required")
		}
	case "email":
		var m map[string]any
		if err := json.Unmarshal([]byte(cfgJSON), &m); err != nil {
			return errors.New("invalid email config")
		}
		for _, k := range []string{"smtp_host", "from", "to"} {
			s, ok := m[k].(string)
			if !ok || strings.TrimSpace(s) == "" {
				return errors.New("email config requires smtp_host, from, and to")
			}
		}
	}
	return nil
}

func chatIDFromConfig(m map[string]any) string {
	v, ok := m["chat_id"]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return fmt.Sprintf("%.0f", x)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

// ListNotificationChannels returns configured notification channels (secrets never exposed).
func ListNotificationChannels(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rows []models.NotificationChannel
		if err := db.Order("sort_order, name").Find(&rows).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out := make([]notificationChannelDTO, 0, len(rows))
		for _, row := range rows {
			out = append(out, toNotificationDTO(row))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

type createNotificationChannelBody struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config"`
	Secret    string         `json:"secret"`
	Enabled   *bool          `json:"enabled"`
	SortOrder int            `json:"sort_order"`
}

// CreateNotificationChannel adds a channel.
func CreateNotificationChannel(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body createNotificationChannelBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Type = strings.ToLower(strings.TrimSpace(body.Type))
		body.Secret = strings.TrimSpace(body.Secret)
		if body.Name == "" || !validNotificationType(body.Type) {
			http.Error(w, "name and valid type required (slack, teams, telegram, email)", http.StatusBadRequest)
			return
		}
		var cfgBytes []byte
		if body.Config == nil {
			cfgBytes = []byte("{}")
		} else {
			var err error
			cfgBytes, err = json.Marshal(body.Config)
			if err != nil {
				http.Error(w, "invalid config", http.StatusBadRequest)
				return
			}
		}
		cfgStr, err := normalizeConfigJSON(string(cfgBytes))
		if err != nil {
			http.Error(w, "invalid config json", http.StatusBadRequest)
			return
		}
		if err := validateMergedState(body.Type, cfgStr, body.Secret); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		en := true
		if body.Enabled != nil {
			en = *body.Enabled
		}
		ch := models.NotificationChannel{
			Name:       body.Name,
			Type:       body.Type,
			ConfigJSON: cfgStr,
			Secret:     body.Secret,
			Enabled:    en,
			SortOrder:  body.SortOrder,
		}
		if err := db.Create(&ch).Error; err != nil {
			http.Error(w, "could not create", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toNotificationDTO(ch))
	}
}

type patchNotificationChannelBody struct {
	Name      *string        `json:"name"`
	Type      *string        `json:"type"`
	Config    map[string]any `json:"config"`
	Secret    *string        `json:"secret"`
	Enabled   *bool          `json:"enabled"`
	SortOrder *int           `json:"sort_order"`
	ClearSecret *bool        `json:"clear_secret"`
}

// PatchNotificationChannel updates a channel.
func PatchNotificationChannel(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "notificationChannelID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body patchNotificationChannelBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var ch models.NotificationChannel
		if err := db.First(&ch, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		typ := ch.Type
		cfgStr := ch.ConfigJSON
		sec := ch.Secret

		updates := map[string]any{}
		if body.Name != nil {
			s := strings.TrimSpace(*body.Name)
			if s == "" {
				http.Error(w, "name empty", http.StatusBadRequest)
				return
			}
			updates["name"] = s
		}
		if body.Type != nil {
			t := strings.ToLower(strings.TrimSpace(*body.Type))
			if !validNotificationType(t) {
				http.Error(w, "invalid type", http.StatusBadRequest)
				return
			}
			updates["type"] = t
			typ = t
		}
		if body.Config != nil {
			b, err := json.Marshal(body.Config)
			if err != nil {
				http.Error(w, "invalid config", http.StatusBadRequest)
				return
			}
			normalized, err := normalizeConfigJSON(string(b))
			if err != nil {
				http.Error(w, "invalid config json", http.StatusBadRequest)
				return
			}
			updates["config_json"] = normalized
			cfgStr = normalized
		}
		if body.Secret != nil {
			s := strings.TrimSpace(*body.Secret)
			if s != "" {
				updates["secret"] = s
				sec = s
			}
		}
		if body.ClearSecret != nil && *body.ClearSecret {
			updates["secret"] = ""
			sec = ""
		}
		if body.Enabled != nil {
			updates["enabled"] = *body.Enabled
		}
		if body.SortOrder != nil {
			updates["sort_order"] = *body.SortOrder
		}

		if len(updates) > 0 {
			if err := validateMergedState(typ, cfgStr, sec); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if len(updates) == 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(toNotificationDTO(ch))
			return
		}
		if err := db.Model(&ch).Updates(updates).Error; err != nil {
			http.Error(w, "update error", http.StatusInternalServerError)
			return
		}
		if err := db.First(&ch, id).Error; err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toNotificationDTO(ch))
	}
}

// DeleteNotificationChannel removes a channel.
func DeleteNotificationChannel(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "notificationChannelID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		res := db.Delete(&models.NotificationChannel{}, id)
		if res.Error != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if res.RowsAffected == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type testNotificationResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// TestNotificationChannel sends a test message through the channel.
func TestNotificationChannel(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUintParam(r, "notificationChannelID")
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var ch models.NotificationChannel
		if err := db.First(&ch, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if !ch.Enabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(testNotificationResult{OK: false, Error: "channel is disabled"})
			return
		}
		if err := notify.SendTest(&ch); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(testNotificationResult{OK: false, Error: err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testNotificationResult{OK: true})
	}
}
