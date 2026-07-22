package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/notifications"
	"github.com/switchboard/switchboard/internal/settings"
)

func toUserDTO(user db.User, roles []db.Role) UserDTO {
	names := make([]string, len(roles))
	ids := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
		ids[i] = role.ID.String()
	}
	dto := UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AuthType:  user.AuthType,
		IsActive:  user.IsActive,
		Roles:     names,
		RoleIDs:   ids,
		CreatedAt: user.CreatedAt,
	}
	if user.DisplayName.Valid {
		dto.DisplayName = user.DisplayName.String
	}
	if user.LastLoginAt.Valid {
		t := user.LastLoginAt.Time
		dto.LastLoginAt = &t
	}
	return dto
}

func hashInviteToken(token string) string {
	return auth.HashToken(token)
}

func inviteURL(cfg config.Config, token string) string {
	base := strings.TrimRight(cfg.AppBaseURL, "/")
	return fmt.Sprintf("%s/invite?token=%s", base, token)
}

func sendInviteEmail(ctx context.Context, q *db.Queries, cfg config.Config, email, inviteLink string, triggeredBy *uuid.UUID) error {
	smtp := settings.ResolveSMTP(ctx, q, cfg)
	html := fmt.Sprintf(
		`<p>You have been invited to join Switchboard.</p><p><a href="%s">Accept invitation and set your password</a></p><p>This link expires in 7 days.</p>`,
		inviteLink,
	)
	return notifications.SendOutbound(ctx, q, smtp, []notifications.MailRecipient{
		notifications.RecipientForEmail(email),
	}, notifications.OutboundOptions{
		EventType:   "invite",
		Subject:     "You're invited to Switchboard",
		HTMLBody:    html,
		PlainBody:   "You have been invited to join Switchboard. Open the invite link to set your password.",
		TriggeredBy: triggeredBy,
	})
}

func (h *Handler) assignRoles(ctx context.Context, userID uuid.UUID, roleIDs []string) error {
	_ = h.queries.SetUserRoles(ctx, userID)
	for _, rid := range roleIDs {
		roleUUID, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		if err := h.queries.AddUserRole(ctx, db.AddUserRoleParams{UserID: userID, RoleID: roleUUID}); err != nil {
			return err
		}
	}
	return nil
}

func parseRoleUUIDs(roleIDs []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(roleIDs))
	for _, rid := range roleIDs {
		id, err := uuid.Parse(rid)
		if err != nil {
			return nil, fmt.Errorf("invalid role id")
		}
		out = append(out, id)
	}
	return out, nil
}

func invitationExpiry() time.Time {
	return time.Now().Add(7 * 24 * time.Hour)
}

func toSessionHistory(rows []db.ListUserLoginHistoryRow) []SessionHistoryDTO {
	out := make([]SessionHistoryDTO, len(rows))
	for i, row := range rows {
		dto := SessionHistoryDTO{
			ID:        row.ID,
			IssuedAt:  row.IssuedAt,
			ExpiresAt: row.ExpiresAt,
			Revoked:   row.Revoked,
		}
		if row.UserAgent.Valid {
			dto.UserAgent = row.UserAgent.String
		}
		if row.IpAddress.Valid {
			dto.IPAddress = row.IpAddress.String
		}
		out[i] = dto
	}
	return out
}
