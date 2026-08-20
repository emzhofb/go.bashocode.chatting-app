package auth

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9._]+$`)

var reservedUsernames = map[string]struct{}{
	"admin": {}, "administrator": {}, "api": {}, "app": {}, "help": {}, "root": {}, "support": {}, "system": {}, "tinode": {},
}

func NormalizeUsername(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) < 4 || len(value) > 32 || !utf8.ValidString(value) || !usernamePattern.MatchString(value) {
		return "", fmt.Errorf("username must be 4-32 lowercase letters, digits, dots, or underscores")
	}
	if value[0] == '.' || value[len(value)-1] == '.' || strings.Contains(value, "..") {
		return "", fmt.Errorf("username cannot start/end with a dot or contain consecutive dots")
	}
	if _, found := reservedUsernames[value]; found {
		return "", fmt.Errorf("username is reserved")
	}
	return value, nil
}

func NormalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 320 || !utf8.ValidString(value) || strings.Count(value, "@") != 1 || strings.ContainsAny(value, " \t\r\n") {
		return "", fmt.Errorf("email is invalid")
	}
	parts := strings.SplitN(value, "@", 2)
	if parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", fmt.Errorf("email is invalid")
	}
	return value, nil
}

func ValidatePassword(value string) error {
	if utf8.RuneCountInString(value) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}
	if len([]byte(value)) > 72 {
		return fmt.Errorf("password must be at most 72 bytes")
	}
	return nil
}

func ValidateDisplayName(value string) error {
	value = strings.TrimSpace(value)
	count := utf8.RuneCountInString(value)
	if count < 1 || count > 80 {
		return fmt.Errorf("display name must be 1-80 characters")
	}
	return nil
}
