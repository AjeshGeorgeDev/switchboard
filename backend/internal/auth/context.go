package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
	rolesKey    contextKey = "roles"
)

func WithUser(ctx context.Context, userID uuid.UUID, username string, roles []string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, usernameKey, username)
	ctx = context.WithValue(ctx, rolesKey, roles)
	return ctx
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(usernameKey).(string)
	return u, ok
}

func RolesFromContext(ctx context.Context) []string {
	roles, ok := ctx.Value(rolesKey).([]string)
	if !ok {
		return nil
	}
	return roles
}
