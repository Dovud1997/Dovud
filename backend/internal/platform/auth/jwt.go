package auth

import (
	"fmt"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	cfg config.AuthConfig
}

func NewTokenService(cfg config.AuthConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

type accessClaims struct {
	TenantID        string   `json:"tenant_id"`
	Roles           []string `json:"roles"`
	Permissions     []string `json:"permissions"`
	DeviceID        string   `json:"device_id,omitempty"`
	IsPlatformAdmin bool     `json:"is_platform_admin"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int64     `json:"expires_in"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	TokenType        string    `json:"token_type"`
}

func (s *TokenService) IssueAccessToken(userID, tenantID uuid.UUID, roles, permissions []string, deviceID string, isPlatformAdmin bool) (string, string, time.Time, error) {
	jti := uuid.NewString()
	exp := time.Now().UTC().Add(s.cfg.AccessTTL())
	claims := accessClaims{
		TenantID:        tenantID.String(),
		Roles:           roles,
		Permissions:     permissions,
		DeviceID:        deviceID,
		IsPlatformAdmin: isPlatformAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   userID.String(),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.AccessSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}
	return signed, jti, exp, nil
}

func (s *TokenService) ParseAccessToken(tokenStr string) (*httpx.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, err
	}
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, err
	}
	return &httpx.TokenClaims{
		UserID:          userID,
		TenantID:        tenantID,
		Roles:           claims.Roles,
		Permissions:     claims.Permissions,
		DeviceID:        claims.DeviceID,
		JTI:             claims.ID,
		IsPlatformAdmin: claims.IsPlatformAdmin,
	}, nil
}

func (s *TokenService) NewRefreshToken() (plain string, expiresAt time.Time, err error) {
	plain = uuid.NewString() + uuid.NewString()
	expiresAt = time.Now().UTC().Add(s.cfg.RefreshTTL())
	return plain, expiresAt, nil
}

func (s *TokenService) AccessTTLSeconds() int64 {
	return int64(s.cfg.AccessTTL().Seconds())
}
