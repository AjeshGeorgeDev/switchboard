package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/db"
)

type SessionService struct {
	queries *db.Queries
	tokens  *TokenService
}

func NewSessionService(queries *db.Queries, tokens *TokenService) *SessionService {
	return &SessionService{queries: queries, tokens: tokens}
}

func (s *SessionService) IssueSession(ctx context.Context, userID uuid.UUID, userAgent, ip string, remember bool) (access, refresh string, accessExp time.Time, refreshTTL time.Duration, err error) {
	refreshTTL = s.tokens.SessionRefreshTTL()
	if remember {
		refreshTTL = s.tokens.RememberRefreshTTL()
	}
	return s.issueSession(ctx, userID, userAgent, ip, refreshTTL)
}

func (s *SessionService) issueSession(ctx context.Context, userID uuid.UUID, userAgent, ip string, refreshTTL time.Duration) (access, refresh string, accessExp time.Time, outTTL time.Duration, err error) {
	access, accessExp, err = s.tokens.CreateAccessToken(userID)
	if err != nil {
		return "", "", time.Time{}, 0, err
	}

	refresh, err = NewRefreshToken()
	if err != nil {
		return "", "", time.Time{}, 0, err
	}

	_, err = s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: HashToken(refresh),
		ExpiresAt: time.Now().Add(refreshTTL),
		UserAgent: pgtype.Text{String: userAgent, Valid: userAgent != ""},
		IpAddress: pgtype.Text{String: ip, Valid: ip != ""},
	})
	if err != nil {
		return "", "", time.Time{}, 0, err
	}

	return access, refresh, accessExp, refreshTTL, nil
}

func (s *SessionService) RotateRefresh(ctx context.Context, refreshToken, userAgent, ip string) (access, newRefresh string, accessExp time.Time, refreshTTL time.Duration, userID uuid.UUID, err error) {
	row, err := s.queries.GetRefreshTokenByHash(ctx, HashToken(refreshToken))
	if err != nil {
		return "", "", time.Time{}, 0, uuid.Nil, err
	}
	if row.ExpiresAt.Before(time.Now()) {
		return "", "", time.Time{}, 0, uuid.Nil, err
	}

	refreshTTL = row.ExpiresAt.Sub(row.IssuedAt)
	if refreshTTL > s.tokens.RememberRefreshTTL() {
		refreshTTL = s.tokens.RememberRefreshTTL()
	}
	if refreshTTL < s.tokens.SessionRefreshTTL() {
		refreshTTL = s.tokens.SessionRefreshTTL()
	}

	_ = s.queries.RevokeRefreshToken(ctx, row.ID)
	access, newRefresh, exp, ttl, err := s.issueSession(ctx, row.UserID, userAgent, ip, refreshTTL)
	return access, newRefresh, exp, ttl, row.UserID, err
}

func (s *SessionService) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return s.queries.RevokeAllUserRefreshTokens(ctx, userID)
}

func SetAuthCookies(w http.ResponseWriter, access, refresh string, accessExp time.Time, refreshTTL time.Duration, remember bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Expires:  accessExp,
	})

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	}
	if remember {
		refreshCookie.Expires = time.Now().Add(refreshTTL)
	}
	http.SetCookie(w, &refreshCookie)
}

func ClearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{"access_token", "refresh_token"} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	}
}

func AccessTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie("access_token"); err == nil && c.Value != "" {
		return c.Value
	}
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func RefreshTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie("refresh_token"); err == nil {
		return c.Value
	}
	return ""
}
