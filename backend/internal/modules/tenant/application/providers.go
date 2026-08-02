package application

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/crypto"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
	"github.com/google/uuid"
)

type ProviderDTO struct {
	Type      string         `json:"type"`
	Driver    string         `json:"driver"`
	IsEnabled bool           `json:"is_enabled"`
	Config    map[string]any `json:"config"`
}

type UpsertProviderInput struct {
	Driver    string         `json:"driver"`
	IsEnabled *bool          `json:"is_enabled"`
	Config    map[string]any `json:"config"`
}

type TestProviderInput struct {
	To string `json:"to"`
}

func (s *TenantService) ListProviders(ctx context.Context, tenantID uuid.UUID) ([]ProviderDTO, error) {
	if s.providers == nil || s.box == nil {
		return nil, apperrors.ErrUnavailable
	}
	rows, err := s.providers.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderDTO, 0, len(rows))
	for _, p := range rows {
		dto, err := s.toProviderDTO(p)
		if err != nil {
			return nil, err
		}
		out = append(out, *dto)
	}
	return out, nil
}

func (s *TenantService) UpsertProvider(ctx context.Context, tenantID uuid.UUID, providerType string, in UpsertProviderInput) (*ProviderDTO, error) {
	if s.providers == nil || s.box == nil {
		return nil, apperrors.ErrUnavailable
	}
	providerType = strings.ToLower(strings.TrimSpace(providerType))
	if !domain.ValidProviderType(providerType) {
		return nil, apperrors.ErrValidation
	}
	driver := strings.ToLower(strings.TrimSpace(in.Driver))
	if driver == "" {
		driver = "log"
	}
	cfgMap := map[string]any{}
	if in.Config != nil {
		cfgMap = in.Config
	}
	// merge with existing secrets when blank placeholders are sent
	if existing, err := s.providers.FindByType(ctx, tenantID, providerType); err == nil {
		prev, _ := s.decryptConfig(existing.ConfigEnc)
		for _, secretKey := range []string{"password", "api_key", "webhook_url", "secret", "service_account_json", "private_key"} {
			v, _ := cfgMap[secretKey].(string)
			if v == "" || v == "********" {
				if prev != nil {
					if pv, ok := prev[secretKey]; ok {
						cfgMap[secretKey] = pv
					}
				}
			}
		}
		if in.IsEnabled == nil {
			en := existing.IsEnabled
			in.IsEnabled = &en
		}
	}
	raw, err := json.Marshal(cfgMap)
	if err != nil {
		return nil, err
	}
	enc, err := s.box.Seal(raw)
	if err != nil {
		return nil, err
	}
	enabled := true
	if in.IsEnabled != nil {
		enabled = *in.IsEnabled
	}
	p := &domain.TenantProvider{
		TenantID: tenantID, Type: providerType, Driver: driver,
		ConfigEnc: enc, IsEnabled: enabled,
	}
	if err := s.providers.Upsert(ctx, p); err != nil {
		return nil, err
	}
	stored, err := s.providers.FindByType(ctx, tenantID, providerType)
	if err != nil {
		return nil, err
	}
	return s.toProviderDTO(*stored)
}

func (s *TenantService) TestProvider(ctx context.Context, tenantID uuid.UUID, providerType string, in TestProviderInput) error {
	providerType = strings.ToLower(strings.TrimSpace(providerType))
	p, err := s.providers.FindByType(ctx, tenantID, providerType)
	if err != nil {
		return err
	}
	if !p.IsEnabled {
		return apperrors.New("PROVIDER_DISABLED", "Provider is disabled", 400)
	}
	cfgMap, err := s.decryptConfig(p.ConfigEnc)
	if err != nil {
		return err
	}
	notifyCfg := s.toNotifyConfig(providerType, p.Driver, cfgMap)
	router := notify.NewRouter(notifyCfg, nil)
	to := strings.TrimSpace(in.To)
	if to == "" {
		to = "test@example.com"
	}
	channel := providerType
	if providerType == domain.ProviderSMTP {
		channel = "email"
	}
	return router.Send(ctx, channel, notify.Message{
		To: to, Subject: "SFA provider test", Body: "Test notification from SFA tenant providers.",
	})
}

// ResolveNotifyConfig merges tenant provider overrides into the base notify config.
func (s *TenantService) ResolveNotifyConfig(ctx context.Context, tenantID uuid.UUID, base config.NotifyConfig) config.NotifyConfig {
	out := base
	if p, err := s.providers.FindByType(ctx, tenantID, domain.ProviderSMTP); err == nil && p.IsEnabled {
		if cfg, err := s.decryptConfig(p.ConfigEnc); err == nil {
			applySMTP(&out, p.Driver, cfg)
		}
	}
	if p, err := s.providers.FindByType(ctx, tenantID, domain.ProviderSMS); err == nil && p.IsEnabled {
		if cfg, err := s.decryptConfig(p.ConfigEnc); err == nil {
			applySMS(&out, p.Driver, cfg)
		}
	}
	if p, err := s.providers.FindByType(ctx, tenantID, domain.ProviderPush); err == nil && p.IsEnabled {
		if cfg, err := s.decryptConfig(p.ConfigEnc); err == nil {
			applyPush(&out, p.Driver, cfg)
		}
	}
	return out
}

func (s *TenantService) toProviderDTO(p domain.TenantProvider) (*ProviderDTO, error) {
	cfg, err := s.decryptConfig(p.ConfigEnc)
	if err != nil {
		cfg = map[string]any{}
	}
	maskSecrets(cfg)
	return &ProviderDTO{
		Type: p.Type, Driver: p.Driver, IsEnabled: p.IsEnabled, Config: cfg,
	}, nil
}

func (s *TenantService) decryptConfig(enc string) (map[string]any, error) {
	raw, err := s.box.Open(enc)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func (s *TenantService) toNotifyConfig(providerType, driver string, cfg map[string]any) config.NotifyConfig {
	out := config.NotifyConfig{}
	switch providerType {
	case domain.ProviderSMTP:
		applySMTP(&out, driver, cfg)
	case domain.ProviderSMS:
		applySMS(&out, driver, cfg)
	case domain.ProviderPush:
		applyPush(&out, driver, cfg)
	}
	return out
}

func applySMTP(out *config.NotifyConfig, driver string, cfg map[string]any) {
	if driver == "" {
		driver = str(cfg, "driver", "log")
	}
	out.Email.Driver = driver
	out.Email.Host = str(cfg, "host", "localhost")
	out.Email.Port = intVal(cfg, "port", 1025)
	out.Email.Username = str(cfg, "username", "")
	out.Email.Password = str(cfg, "password", "")
	out.Email.From = str(cfg, "from", "noreply@sfa.local")
	out.Email.FileDir = str(cfg, "file_dir", "./storage/mail")
}

func applySMS(out *config.NotifyConfig, driver string, cfg map[string]any) {
	if driver == "" {
		driver = str(cfg, "driver", "log")
	}
	out.SMS.Driver = driver
	out.SMS.WebhookURL = str(cfg, "webhook_url", "")
}

func applyPush(out *config.NotifyConfig, driver string, cfg map[string]any) {
	if driver == "" {
		driver = str(cfg, "driver", "log")
	}
	out.Push.Driver = driver
	out.Push.WebhookURL = str(cfg, "webhook_url", "")
	out.Push.ProjectID = str(cfg, "project_id", "")
	out.Push.CredentialsJSON = str(cfg, "service_account_json", "")
	if out.Push.CredentialsJSON == "" {
		// allow nested SA fields stored flattened
		if email := str(cfg, "client_email", ""); email != "" && str(cfg, "private_key", "") != "" {
			raw, _ := json.Marshal(map[string]string{
				"client_email": email,
				"private_key":  str(cfg, "private_key", ""),
				"project_id":   str(cfg, "project_id", ""),
			})
			out.Push.CredentialsJSON = string(raw)
		}
	}
}

func maskSecrets(cfg map[string]any) {
	for _, k := range []string{"password", "api_key", "secret", "service_account_json", "private_key"} {
		if v, ok := cfg[k].(string); ok && v != "" {
			cfg[k] = "********"
		}
	}
}

func str(m map[string]any, key, fallback string) string {
	if v, ok := m[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func intVal(m map[string]any, key string, fallback int) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	default:
		return fallback
	}
}

// Ensure box is always usable in tests via WithCrypto.
func EnsureProviderBox(passphrase string) (*crypto.SecretBox, error) {
	return crypto.NewSecretBox(passphrase)
}
