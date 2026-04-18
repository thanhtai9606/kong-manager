package notify

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/kong/kong-manager/internal/models"
)

const testMessage = "Kong Manager: test notification"

var httpClient = &http.Client{Timeout: 20 * time.Second}

// SendTest delivers a single test message using the channel configuration.
func SendTest(ch *models.NotificationChannel) error {
	if ch == nil {
		return errors.New("channel is nil")
	}
	switch strings.ToLower(strings.TrimSpace(ch.Type)) {
	case "slack", "teams":
		return sendWebhookJSON(ch.Secret, map[string]string{"text": testMessage})
	case "telegram":
		return sendTelegram(ch)
	case "email":
		return sendEmailTest(ch)
	default:
		return fmt.Errorf("unsupported channel type %q", ch.Type)
	}
}

func sendWebhookJSON(webhookURL string, payload any) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return errors.New("webhook URL is empty")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpClient.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}
	return nil
}

func sendTelegram(ch *models.NotificationChannel) error {
	token := strings.TrimSpace(ch.Secret)
	if token == "" {
		return errors.New("bot token is empty")
	}
	var m map[string]any
	if strings.TrimSpace(ch.ConfigJSON) != "" {
		if err := json.Unmarshal([]byte(ch.ConfigJSON), &m); err != nil {
			return fmt.Errorf("config json: %w", err)
		}
	}
	chatID := telegramChatIDFromMap(m)
	if chatID == "" {
		return errors.New("chat_id missing in config")
	}
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	return sendWebhookJSON(u, map[string]string{
		"chat_id": chatID,
		"text":    testMessage,
	})
}

func telegramChatIDFromMap(m map[string]any) string {
	if m == nil {
		return ""
	}
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

type emailCfg struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`
	SMTPUser string `json:"smtp_user"`
	From     string `json:"from"`
	To       string `json:"to"`
	UseTLS   bool   `json:"use_tls"`
}

func sendEmailTest(ch *models.NotificationChannel) error {
	var cfg emailCfg
	if err := json.Unmarshal([]byte(strings.TrimSpace(ch.ConfigJSON)), &cfg); err != nil {
		return fmt.Errorf("config json: %w", err)
	}
	cfg.SMTPHost = strings.TrimSpace(cfg.SMTPHost)
	cfg.SMTPPort = strings.TrimSpace(cfg.SMTPPort)
	cfg.SMTPUser = strings.TrimSpace(cfg.SMTPUser)
	cfg.From = strings.TrimSpace(cfg.From)
	cfg.To = strings.TrimSpace(cfg.To)
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = "587"
	}
	if cfg.SMTPHost == "" || cfg.From == "" || cfg.To == "" {
		return errors.New("smtp_host, from, and to are required in config")
	}
	pass := strings.TrimSpace(ch.Secret)
	if pass == "" {
		return errors.New("smtp password (secret) is empty")
	}
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)
	subject := "Kong Manager test notification"
	body := "This is a test message from Kong Manager.\r\n"
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.From, cfg.To, subject, body))
	auth := smtp.PlainAuth("", cfg.SMTPUser, pass, cfg.SMTPHost)
	if cfg.UseTLS {
		tlsCfg := &tls.Config{ServerName: cfg.SMTPHost, MinVersion: tls.VersionTLS12}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer c.Close()
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
		if err := c.Mail(cfg.From); err != nil {
			return err
		}
		if err := c.Rcpt(cfg.To); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return c.Quit()
	}
	return smtp.SendMail(addr, auth, cfg.From, []string{cfg.To}, msg)
}
