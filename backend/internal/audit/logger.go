package audit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

type Logger struct {
	queries *db.Queries
}

func New(queries *db.Queries) *Logger {
	return &Logger{queries: queries}
}

func (l *Logger) Log(ctx context.Context, action, resourceType, resourceID string, details map[string]interface{}) {
	if l == nil || l.queries == nil {
		return
	}
	var actorID pgtype.UUID
	actorUsername := ""
	if id, ok := auth.UserIDFromContext(ctx); ok {
		actorID = pgtype.UUID{Bytes: id, Valid: true}
	}
	if name, ok := auth.UsernameFromContext(ctx); ok {
		actorUsername = name
	}
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}
	ip := ""
	if req, ok := ctx.Value(clientIPKey{}).(string); ok {
		ip = req
	}
	_, _ = l.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ActorID:        actorID,
		ActorUsername:  pgtype.Text{String: actorUsername, Valid: actorUsername != ""},
		Action:         action,
		ResourceType:   pgtype.Text{String: resourceType, Valid: resourceType != ""},
		ResourceID:     pgtype.Text{String: resourceID, Valid: resourceID != ""},
		Details:        detailsJSON,
		IpAddress:      pgtype.Text{String: ip, Valid: ip != ""},
	})
}

func (l *Logger) LogAuth(ctx context.Context, actorID uuid.UUID, actorUsername, action string, r *http.Request, details map[string]interface{}) {
	if l == nil || l.queries == nil {
		return
	}
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}
	ip := ""
	ua := ""
	if r != nil {
		ip = r.RemoteAddr
		ua = r.UserAgent()
		if details == nil {
			details = map[string]interface{}{}
		}
		details["user_agent"] = ua
		detailsJSON, _ = json.Marshal(details)
	}
	_, _ = l.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ActorID:        pgtype.UUID{Bytes: actorID, Valid: actorID != uuid.Nil},
		ActorUsername:  pgtype.Text{String: actorUsername, Valid: actorUsername != ""},
		Action:         action,
		ResourceType:   pgtype.Text{String: "auth", Valid: true},
		ResourceID:     pgtype.Text{String: actorUsername, Valid: actorUsername != ""},
		Details:        detailsJSON,
		IpAddress:      pgtype.Text{String: ip, Valid: ip != ""},
	})
}

type clientIPKey struct{}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}
