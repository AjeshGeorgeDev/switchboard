package users

import (
	"time"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name,omitempty"`
	AuthType    string     `json:"auth_type"`
	IsActive    bool       `json:"is_active"`
	Roles       []string   `json:"roles"`
	RoleIDs     []string   `json:"role_ids"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type InvitationDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	RoleIDs   []string  `json:"role_ids"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionHistoryDTO struct {
	ID        uuid.UUID `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}
