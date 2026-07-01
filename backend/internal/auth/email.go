package auth

import "strings"

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func EmailAsUsername(email string) string {
	return NormalizeEmail(email)
}
