package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kong/kong-manager/internal/config"
)

// Claims is embedded in JWT access tokens.
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"sub"`
	jwt.RegisteredClaims
}

// OIDCStateClaims is signed into the OAuth state parameter for CSRF protection.
type OIDCStateClaims struct {
	Slug  string `json:"slg"`
	Pid   uint   `json:"pid"`
	Nonce string `json:"nce"`
	jwt.RegisteredClaims
}

// SignOIDCState returns a short-lived JWT for the OIDC authorize redirect.
func (s *Service) SignOIDCState(slug string, providerID uint, nonce string) (string, error) {
	now := time.Now()
	c := OIDCStateClaims{
		Slug:  slug,
		Pid:   providerID,
		Nonce: nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "kong-manager-oidc-state",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	return t.SignedString([]byte(s.cfg.JWTSecret))
}

// ParseOIDCState validates state from the OAuth callback.
func (s *Service) ParseOIDCState(raw string) (*OIDCStateClaims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &OIDCStateClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*OIDCStateClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid state")
	}
	return claims, nil
}

// TokenPair is returned after successful login (MVP: access token only).
type TokenPair struct {
	AccessToken string `json:"token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// Service issues and parses JWTs.
type Service struct {
	cfg *config.Config
}

// NewService constructs a JWT service.
func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// IssueAccessToken creates a signed JWT for the given user.
func (s *Service) IssueAccessToken(userID uint, username string) (*TokenPair, error) {
	now := time.Now()
	exp := now.Add(s.cfg.JWTTTL)
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kong-manager",
			Subject:   username,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken: signed,
		ExpiresIn:   int64(s.cfg.JWTTTL.Seconds()),
		TokenType:   "Bearer",
	}, nil
}

// Parse validates a Bearer token and returns claims.
func (s *Service) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
