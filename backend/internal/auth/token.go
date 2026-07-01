package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/config"
)

type TokenService struct {
	secret      []byte
	accessTTL   time.Duration
	sessionTTL  time.Duration
	rememberTTL time.Duration
}

type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewTokenService(cfg config.Config) *TokenService {
	return &TokenService{
		secret:      []byte(cfg.JWTSecret),
		accessTTL:   cfg.JWTAccessTTL,
		sessionTTL:  cfg.JWTSessionRefreshTTL,
		rememberTTL: cfg.JWTRememberRefreshTTL,
	}
}

func (s *TokenService) CreateAccessToken(userID uuid.UUID) (string, time.Time, error) {
	expires := time.Now().Add(s.accessTTL)
	claims := AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	return signed, expires, err
}

func (s *TokenService) ParseAccessToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}

func (s *TokenService) SessionRefreshTTL() time.Duration {
	return s.sessionTTL
}

func (s *TokenService) RememberRefreshTTL() time.Duration {
	return s.rememberTTL
}

func (s *TokenService) RefreshTTL() time.Duration {
	return s.sessionTTL
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
