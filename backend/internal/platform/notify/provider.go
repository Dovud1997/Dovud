package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
)

type Message struct {
	Channel string
	To      string
	Subject string
	Body    string
	Data    map[string]any
}

type Provider interface {
	Name() string
	Send(ctx context.Context, msg Message) error
}

type Router struct {
	Email Provider
	SMS   Provider
	Push  Provider
	Log   *slog.Logger
}

func (r *Router) Send(ctx context.Context, channel string, msg Message) error {
	msg.Channel = channel
	var p Provider
	switch channel {
	case "email":
		p = r.Email
	case "sms":
		p = r.SMS
	case "push":
		p = r.Push
	default:
		return fmt.Errorf("unsupported channel %q", channel)
	}
	if p == nil {
		return fmt.Errorf("provider not configured for %s", channel)
	}
	if err := p.Send(ctx, msg); err != nil {
		return err
	}
	if r.Log != nil {
		r.Log.Info("notification sent", "provider", p.Name(), "channel", channel, "to", msg.To)
	}
	return nil
}

func NewRouter(cfg config.NotifyConfig, log *slog.Logger) *Router {
	if log == nil {
		log = slog.Default()
	}
	r := &Router{Log: log}
	switch strings.ToLower(strings.TrimSpace(cfg.Email.Driver)) {
	case "smtp":
		r.Email = &SMTPEmail{cfg: cfg.Email, log: log}
	case "file":
		r.Email = &FileEmail{dir: cfg.Email.FileDir, log: log}
	default:
		r.Email = &LogEmail{log: log}
	}
	switch strings.ToLower(strings.TrimSpace(cfg.SMS.Driver)) {
	case "http", "webhook":
		r.SMS = &HTTPWebhookSMS{url: cfg.SMS.WebhookURL, log: log, client: &http.Client{Timeout: 10 * time.Second}}
	default:
		r.SMS = &LogSMS{log: log}
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Push.Driver)) {
	case "http", "webhook":
		r.Push = &HTTPWebhookPush{url: cfg.Push.WebhookURL, log: log, client: &http.Client{Timeout: 10 * time.Second}}
	case "fcm":
		p, err := NewFCMPush(cfg.Push, log)
		if err != nil {
			log.Warn("fcm push misconfigured; falling back to log", "error", err)
			r.Push = &LogPush{log: log}
		} else {
			r.Push = p
		}
	default:
		r.Push = &LogPush{log: log}
	}
	return r
}

type LogEmail struct{ log *slog.Logger }

func (p *LogEmail) Name() string { return "log-email" }
func (p *LogEmail) Send(ctx context.Context, msg Message) error {
	_ = ctx
	p.log.Info("email (log)", "to", msg.To, "subject", msg.Subject, "body", msg.Body)
	return nil
}

type FileEmail struct {
	dir string
	log *slog.Logger
}

func (p *FileEmail) Name() string { return "file-email" }
func (p *FileEmail) Send(ctx context.Context, msg Message) error {
	_ = ctx
	dir := p.dir
	if dir == "" {
		dir = "./storage/mail"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("%d_%s.eml", time.Now().UnixNano(), sanitize(msg.To))
	content := fmt.Sprintf("To: %s\nSubject: %s\nDate: %s\n\n%s\n",
		msg.To, msg.Subject, time.Now().UTC().Format(time.RFC1123Z), msg.Body)
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

type SMTPEmail struct {
	cfg config.EmailConfig
	log *slog.Logger
}

func (p *SMTPEmail) Name() string { return "smtp" }
func (p *SMTPEmail) Send(ctx context.Context, msg Message) error {
	_ = ctx
	from := p.cfg.From
	if from == "" {
		from = "noreply@sfa.local"
	}
	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)
	body := strings.Join([]string{
		"From: " + from,
		"To: " + msg.To,
		"Subject: " + msg.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		msg.Body,
	}, "\r\n")
	var auth smtp.Auth
	if p.cfg.Username != "" {
		auth = smtp.PlainAuth("", p.cfg.Username, p.cfg.Password, p.cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{msg.To}, []byte(body))
}

type LogSMS struct{ log *slog.Logger }

func (p *LogSMS) Name() string { return "log-sms" }
func (p *LogSMS) Send(ctx context.Context, msg Message) error {
	_ = ctx
	p.log.Info("sms (log)", "to", msg.To, "body", msg.Body)
	return nil
}

type HTTPWebhookSMS struct {
	url    string
	log    *slog.Logger
	client *http.Client
}

func (p *HTTPWebhookSMS) Name() string { return "http-sms" }
func (p *HTTPWebhookSMS) Send(ctx context.Context, msg Message) error {
	if strings.TrimSpace(p.url) == "" {
		return fmt.Errorf("sms webhook url not configured")
	}
	payload, _ := json.Marshal(map[string]any{
		"channel": "sms", "to": msg.To, "body": msg.Body, "data": msg.Data,
	})
	return postJSON(ctx, p.client, p.url, payload)
}

type LogPush struct{ log *slog.Logger }

func (p *LogPush) Name() string { return "log-push" }
func (p *LogPush) Send(ctx context.Context, msg Message) error {
	_ = ctx
	p.log.Info("push (log)", "to", msg.To, "subject", msg.Subject, "body", msg.Body, "data", msg.Data)
	return nil
}

type HTTPWebhookPush struct {
	url    string
	log    *slog.Logger
	client *http.Client
}

func (p *HTTPWebhookPush) Name() string { return "http-push" }
func (p *HTTPWebhookPush) Send(ctx context.Context, msg Message) error {
	if strings.TrimSpace(p.url) == "" {
		return fmt.Errorf("push webhook url not configured")
	}
	payload, _ := json.Marshal(map[string]any{
		"channel": "push", "to": msg.To, "title": msg.Subject, "body": msg.Body, "data": msg.Data,
	})
	return postJSON(ctx, p.client, p.url, payload)
}

func postJSON(ctx context.Context, client *http.Client, url string, payload []byte) error {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "@", "_at_")
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, s)
	return s
}
