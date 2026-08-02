package notify

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultOAuthTokenURL = "https://oauth2.googleapis.com/token"
	fcmScope             = "https://www.googleapis.com/auth/firebase.messaging"
)

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	ProjectID   string `json:"project_id"`
	TokenURI    string `json:"token_uri"`
}

// FCMPush sends via Firebase Cloud Messaging HTTP v1.
type FCMPush struct {
	projectID string
	email     string
	key       *rsa.PrivateKey
	client    *http.Client
	tokenURL  string
	sendURL   string
	log       *slog.Logger

	mu    sync.Mutex
	token string
	exp   time.Time
}

func NewFCMPush(cfg config.PushConfig, log *slog.Logger) (*FCMPush, error) {
	if log == nil {
		log = slog.Default()
	}
	raw := strings.TrimSpace(cfg.CredentialsJSON)
	if raw == "" && strings.TrimSpace(cfg.CredentialsFile) != "" {
		b, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("read fcm credentials file: %w", err)
		}
		raw = string(b)
	}
	if raw == "" {
		return nil, fmt.Errorf("fcm credentials_json not configured")
	}
	var sa serviceAccount
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return nil, fmt.Errorf("parse fcm credentials: %w", err)
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, fmt.Errorf("fcm credentials missing client_email/private_key")
	}
	projectID := strings.TrimSpace(cfg.ProjectID)
	if projectID == "" {
		projectID = strings.TrimSpace(sa.ProjectID)
	}
	if projectID == "" {
		return nil, fmt.Errorf("fcm project_id not configured")
	}
	key, err := parseRSAPrivateKey(sa.PrivateKey)
	if err != nil {
		return nil, err
	}
	tokenURL := defaultOAuthTokenURL
	if strings.TrimSpace(sa.TokenURI) != "" {
		tokenURL = strings.TrimSpace(sa.TokenURI)
	}
	return &FCMPush{
		projectID: projectID,
		email:     sa.ClientEmail,
		key:       key,
		client:    &http.Client{Timeout: 15 * time.Second},
		tokenURL:  tokenURL,
		sendURL:   fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", url.PathEscape(projectID)),
		log:       log,
	}, nil
}

// WithEndpoints overrides OAuth/FCM endpoints (tests).
func (p *FCMPush) WithEndpoints(tokenURL, sendURL string) *FCMPush {
	if tokenURL != "" {
		p.tokenURL = tokenURL
	}
	if sendURL != "" {
		p.sendURL = sendURL
	}
	return p
}

func (p *FCMPush) WithHTTPClient(c *http.Client) *FCMPush {
	if c != nil {
		p.client = c
	}
	return p
}

func (p *FCMPush) Name() string { return "fcm-push" }

func (p *FCMPush) Send(ctx context.Context, msg Message) error {
	token := strings.TrimSpace(msg.To)
	if token == "" {
		return fmt.Errorf("push token is empty")
	}
	if strings.HasPrefix(token, "stub-push-") {
		p.log.Info("skip stub push token", "to", token)
		return nil
	}
	access, err := p.accessToken(ctx)
	if err != nil {
		return err
	}
	data := map[string]string{}
	for k, v := range msg.Data {
		data[k] = fmt.Sprint(v)
	}
	payload := map[string]any{
		"message": map[string]any{
			"token": token,
			"notification": map[string]string{
				"title": msg.Subject,
				"body":  msg.Body,
			},
			"data": data,
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.sendURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("fcm status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (p *FCMPush) accessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.token != "" && time.Now().Before(p.exp.Add(-30*time.Second)) {
		return p.token, nil
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   p.email,
		"scope": fcmScope,
		"aud":   p.tokenURL,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	assertion, err := tok.SignedString(p.key)
	if err != nil {
		return "", fmt.Errorf("sign fcm jwt: %w", err)
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth token status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode oauth token: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("oauth token empty")
	}
	exp := time.Hour
	if out.ExpiresIn > 0 {
		exp = time.Duration(out.ExpiresIn) * time.Second
	}
	p.token = out.AccessToken
	p.exp = time.Now().Add(exp)
	return p.token, nil
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid fcm private_key pem")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("fcm private_key is not RSA")
		}
		return rsaKey, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
