package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	TenantID string `json:"tenant"`
	jwt.RegisteredClaims
}
type RefreshTokenClaims struct {
	jwt.RegisteredClaims
}
type TokenGenerator struct {
	accessSecretKey  []byte
	refreshSecretKey []byte
	issuer           string
	accessExpiry     time.Duration
}

type Config struct {
	AccessSecret  string
	RefreshSecret string
	Issuer        string
	AccessExpiry  time.Duration
}

func NewTokenGenerator(cfg Config) (*TokenGenerator, error) {
	if cfg.AccessSecret == "" || cfg.RefreshSecret == "" {
		return nil, fmt.Errorf("segredos JWT não podem estar vazios")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("emissor (issuer) JWT não pode estar vazio")
	}
	if cfg.AccessExpiry <= 0 {
		return nil, fmt.Errorf("expiração do token deve ser positiva")
	}

	return &TokenGenerator{
		accessSecretKey:  []byte(cfg.AccessSecret),
		refreshSecretKey: []byte(cfg.RefreshSecret),
		issuer:           cfg.Issuer,
		accessExpiry:     cfg.AccessExpiry,
	}, nil
}
func (tg *TokenGenerator) GenerateAccessToken(userID uuid.UUID, tenantID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(tg.accessExpiry)

	claims := &AccessTokenClaims{
		TenantID: tenantID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    tg.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tg.accessSecretKey)
	if err != nil {
		return "", fmt.Errorf("erro ao assinar o access token: %w", err)
	}

	return tokenString, nil
}

func (tg *TokenGenerator) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := &RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  userID.String(),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer:   tg.issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tg.refreshSecretKey)
	if err != nil {
		return "", fmt.Errorf("erro ao assinar o refresh token: %w", err)
	}
	return tokenString, nil
}
