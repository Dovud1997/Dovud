package config

import (
	"fmt"
	"os"
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
	AccessSecret      string `yaml:"access_secret"`
	RefreshSecret     string `yaml:"refresh_secret"`
	AccessTTLMinutes  int    `yaml:"access_ttl_minutes"`
	RefreshTTLDays    int    `yaml:"refresh_ttl_days"`
	Issuer            string `yaml:"issuer"`
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

	return &cfg, nil
}
