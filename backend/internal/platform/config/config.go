package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	Minio    MinioConfig    `yaml:"minio"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Notify   NotifyConfig   `yaml:"notify"`
	Security SecurityConfig `yaml:"security"`
}

type SecurityConfig struct {
	CORSOrigins           string `yaml:"cors_origins"`
	BodyLimitMB           int    `yaml:"body_limit_mb"`
	RateLimitMax          int    `yaml:"rate_limit_max"`
	RateLimitWindowSec    int    `yaml:"rate_limit_window_sec"`
	LoginMaxAttempts      int    `yaml:"login_max_attempts"`
	LoginWindowMinutes    int    `yaml:"login_window_minutes"`
	LoginLockoutMinutes   int    `yaml:"login_lockout_minutes"`
}

func (s SecurityConfig) BodyLimitBytes() int {
	mb := s.BodyLimitMB
	if mb <= 0 {
		mb = 10
	}
	return mb * 1024 * 1024
}

func (s SecurityConfig) RateLimitWindow() time.Duration {
	sec := s.RateLimitWindowSec
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

type NotifyConfig struct {
	Email EmailConfig `yaml:"email"`
	SMS   SMSConfig   `yaml:"sms"`
	Push  PushConfig  `yaml:"push"`
}

type EmailConfig struct {
	Driver   string `yaml:"driver"` // log | file | smtp
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FileDir  string `yaml:"file_dir"`
}

type SMSConfig struct {
	Driver     string `yaml:"driver"` // log | http
	WebhookURL string `yaml:"webhook_url"`
}

type PushConfig struct {
	Driver     string `yaml:"driver"` // log | http (FCM later)
	WebhookURL string `yaml:"webhook_url"`
}

type AppConfig struct {
	Name          string `yaml:"name"`
	Env           string `yaml:"env"`
	HTTPAddr      string `yaml:"http_addr"`
	PublicBaseURL string `yaml:"public_base_url"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AuthConfig struct {
	AccessSecret     string `yaml:"access_secret"`
	RefreshSecret    string `yaml:"refresh_secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLDays   int    `yaml:"refresh_ttl_days"`
	Issuer           string `yaml:"issuer"`
}

func (a AuthConfig) AccessTTL() time.Duration {
	return time.Duration(a.AccessTTLMinutes) * time.Minute
}

func (a AuthConfig) RefreshTTL() time.Duration {
	return time.Duration(a.RefreshTTLDays) * 24 * time.Hour
}

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
	// Driver: "minio" (default), "local", or "auto" (try minio, fall back to local).
	Driver    string `yaml:"driver"`
	LocalPath string `yaml:"local_path"`
}

type RabbitMQConfig struct {
	URL string `yaml:"url"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("SFA_CONFIG")
	}
	if path == "" {
		path = "configs/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if dsn := os.Getenv("SFA_DATABASE_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if addr := os.Getenv("SFA_HTTP_ADDR"); addr != "" {
		cfg.App.HTTPAddr = addr
	}
	if secret := os.Getenv("SFA_ACCESS_SECRET"); secret != "" {
		cfg.Auth.AccessSecret = secret
	}
	if secret := os.Getenv("SFA_REFRESH_SECRET"); secret != "" {
		cfg.Auth.RefreshSecret = secret
	}
	if redisAddr := os.Getenv("SFA_REDIS_ADDR"); redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	}
	if rabbitURL := os.Getenv("SFA_RABBITMQ_URL"); rabbitURL != "" {
		cfg.RabbitMQ.URL = rabbitURL
	}
	if minioEndpoint := os.Getenv("SFA_MINIO_ENDPOINT"); minioEndpoint != "" {
		cfg.Minio.Endpoint = minioEndpoint
	}
	if minioAccess := os.Getenv("SFA_MINIO_ACCESS_KEY"); minioAccess != "" {
		cfg.Minio.AccessKey = minioAccess
	}
	if minioSecret := os.Getenv("SFA_MINIO_SECRET_KEY"); minioSecret != "" {
		cfg.Minio.SecretKey = minioSecret
	}
	if minioBucket := os.Getenv("SFA_MINIO_BUCKET"); minioBucket != "" {
		cfg.Minio.Bucket = minioBucket
	}
	if v := os.Getenv("SFA_STORAGE_DRIVER"); v != "" {
		cfg.Minio.Driver = v
	}
	if v := os.Getenv("SFA_EMAIL_DRIVER"); v != "" {
		cfg.Notify.Email.Driver = v
	}
	if v := os.Getenv("SFA_SMTP_HOST"); v != "" {
		cfg.Notify.Email.Host = v
	}
	if v := os.Getenv("SFA_SMTP_PORT"); v != "" {
		var port int
		_, _ = fmt.Sscanf(v, "%d", &port)
		if port > 0 {
			cfg.Notify.Email.Port = port
		}
	}
	if v := os.Getenv("SFA_SMTP_FROM"); v != "" {
		cfg.Notify.Email.From = v
	}
	if v := os.Getenv("SFA_SMS_DRIVER"); v != "" {
		cfg.Notify.SMS.Driver = v
	}
	if v := os.Getenv("SFA_SMS_WEBHOOK_URL"); v != "" {
		cfg.Notify.SMS.WebhookURL = v
	}
	if v := os.Getenv("SFA_PUSH_DRIVER"); v != "" {
		cfg.Notify.Push.Driver = v
	}
	if v := os.Getenv("SFA_PUSH_WEBHOOK_URL"); v != "" {
		cfg.Notify.Push.WebhookURL = v
	}
	if v := os.Getenv("SFA_CORS_ORIGINS"); v != "" {
		cfg.Security.CORSOrigins = v
	}
	if v := os.Getenv("SFA_RATE_LIMIT_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Security.RateLimitMax = n
		}
	}

	if cfg.Notify.Email.Driver == "" {
		cfg.Notify.Email.Driver = "log"
	}
	if cfg.Notify.Email.Port == 0 {
		cfg.Notify.Email.Port = 1025
	}
	if cfg.Notify.Email.Host == "" {
		cfg.Notify.Email.Host = "localhost"
	}
	if cfg.Notify.Email.From == "" {
		cfg.Notify.Email.From = "noreply@sfa.local"
	}
	if cfg.Notify.Email.FileDir == "" {
		cfg.Notify.Email.FileDir = "./storage/mail"
	}
	if cfg.Notify.SMS.Driver == "" {
		cfg.Notify.SMS.Driver = "log"
	}
	if cfg.Notify.Push.Driver == "" {
		cfg.Notify.Push.Driver = "log"
	}
	if strings.TrimSpace(cfg.Security.CORSOrigins) == "" {
		cfg.Security.CORSOrigins = "*"
	}
	if cfg.Security.BodyLimitMB <= 0 {
		cfg.Security.BodyLimitMB = 10
	}
	if cfg.Security.RateLimitMax <= 0 {
		cfg.Security.RateLimitMax = 120
	}
	if cfg.Security.RateLimitWindowSec <= 0 {
		cfg.Security.RateLimitWindowSec = 60
	}
	if cfg.Security.LoginMaxAttempts <= 0 {
		cfg.Security.LoginMaxAttempts = 5
	}
	if cfg.Security.LoginWindowMinutes <= 0 {
		cfg.Security.LoginWindowMinutes = 15
	}
	if cfg.Security.LoginLockoutMinutes <= 0 {
		cfg.Security.LoginLockoutMinutes = 15
	}

	return &cfg, nil
}
