package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Provisioner interface {
	ProvisionUser(ctx context.Context, userID, displayName string) error
}

type Service struct {
	DB              *sql.DB
	Provisioner     Provisioner
	BcryptCost      int
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type User struct {
	ID           string  `json:"id"`
	OpenIMUserID string  `json:"openim_user_id"`
	Username     string  `json:"username"`
	Email        string  `json:"-"`
	DisplayName  string  `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	Status       string  `json:"-"`
	Role         string  `json:"-"`
}

type Session struct {
	ID               string    `json:"-"`
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Login    string
	Password string
}

func (s Service) Register(ctx context.Context, input RegisterInput, userAgent string, ip net.IP) (User, Session, error) {
	username, err := NormalizeUsername(input.Username)
	if err != nil {
		return User{}, Session{}, fieldError("username", err)
	}
	email, err := NormalizeEmail(input.Email)
	if err != nil {
		return User{}, Session{}, fieldError("email", err)
	}
	if err := ValidatePassword(input.Password); err != nil {
		return User{}, Session{}, fieldError("password", err)
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if err := ValidateDisplayName(displayName); err != nil {
		return User{}, Session{}, fieldError("display_name", err)
	}
	if s.DB == nil || s.Provisioner == nil {
		return User{}, Session{}, errors.New("auth service is not configured")
	}
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1 OR email = $2)`, username, email).Scan(&exists); err != nil {
		return User{}, Session{}, fmt.Errorf("check registration: %w", err)
	}
	if exists {
		return User{}, Session{}, errors.New("username or email is already registered")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.BcryptCost)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("hash password: %w", err)
	}
	userID, err := NewUUID()
	if err != nil {
		return User{}, Session{}, err
	}
	openIMUserID := openIMUserID(userID)
	// Provision before committing the app row. The OpenIM operation is idempotent:
	// a retry first checks the existing user and does not create a second mapping.
	if err := s.Provisioner.ProvisionUser(ctx, openIMUserID, displayName); err != nil {
		return User{}, Session{}, fmt.Errorf("provision OpenIM user: %w", err)
	}

	now := time.Now().UTC()
	user := User{ID: userID, OpenIMUserID: openIMUserID, Username: username, Email: email, DisplayName: displayName, Status: "active", Role: "user"}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("begin register: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO users (id, openim_user_id, openim_provisioned_at, username, email, password_hash, display_name, status, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', 'user', $8, $8)`, user.ID, user.OpenIMUserID, now, user.Username, user.Email, string(passwordHash), user.DisplayName, now)
	if err != nil {
		return User{}, Session{}, normalizeDBError(err)
	}
	session, err := newSession(user.ID, now, s.AccessTokenTTL, s.RefreshTokenTTL)
	if err != nil {
		return User{}, Session{}, err
	}
	if err := insertSession(ctx, tx, user.ID, session, userAgent, ip, now); err != nil {
		return User{}, Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, Session{}, fmt.Errorf("commit register: %w", err)
	}
	return user, session, nil
}

func openIMUserID(appUserID string) string {
	return "app_" + strings.ReplaceAll(appUserID, "-", "")
}

func (s Service) Login(ctx context.Context, input LoginInput, userAgent string, ip net.IP) (User, Session, error) {
	login := strings.ToLower(strings.TrimSpace(input.Login))
	var user User
	var passwordHash string
	var provisionedAt sql.NullTime
	err := s.DB.QueryRowContext(ctx, `SELECT id, openim_user_id, username, email, password_hash, display_name, avatar_url, status, role, openim_provisioned_at FROM users WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`, login).Scan(&user.ID, &user.OpenIMUserID, &user.Username, &user.Email, &passwordHash, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role, &provisionedAt)
	if err != nil || user.Status != "active" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		return User{}, Session{}, ErrInvalidCredentials
	}
	if !provisionedAt.Valid {
		if s.Provisioner == nil {
			return User{}, Session{}, errors.New("OpenIM provisioner is not configured")
		}
		if err := s.Provisioner.ProvisionUser(ctx, user.OpenIMUserID, user.DisplayName); err != nil {
			return User{}, Session{}, fmt.Errorf("provision OpenIM user: %w", err)
		}
		if _, err := s.DB.ExecContext(ctx, `UPDATE users SET openim_provisioned_at = now(), updated_at = now() WHERE id = $1 AND openim_provisioned_at IS NULL`, user.ID); err != nil {
			return User{}, Session{}, fmt.Errorf("store OpenIM provisioning: %w", err)
		}
	}
	now := time.Now().UTC()
	session, err := newSession(user.ID, now, s.AccessTokenTTL, s.RefreshTokenTTL)
	if err != nil {
		return User{}, Session{}, err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("begin login: %w", err)
	}
	defer tx.Rollback()
	if err := insertSession(ctx, tx, user.ID, session, userAgent, ip, now); err != nil {
		return User{}, Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, Session{}, fmt.Errorf("commit login: %w", err)
	}
	return user, session, nil
}

func (s Service) SearchUsers(ctx context.Context, query, excludeUserID string, limit int) ([]User, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" || s.DB == nil {
		return []User{}, nil
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := s.DB.QueryContext(ctx, `SELECT id, openim_user_id, username, display_name, avatar_url FROM users WHERE deleted_at IS NULL AND status = 'active' AND id <> $2 AND (username ILIKE $1 OR email ILIKE $1 OR display_name ILIKE $1) ORDER BY username LIMIT $3`, pattern, excludeUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()
	users := make([]User, 0, limit)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.OpenIMUserID, &user.Username, &user.DisplayName, &user.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan user search result: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user search results: %w", err)
	}
	return users, nil
}

func (s Service) Refresh(ctx context.Context, refreshToken, userAgent string, ip net.IP) (User, Session, error) {
	if refreshToken == "" {
		return User{}, Session{}, ErrInvalidCredentials
	}
	now := time.Now().UTC()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback()
	var oldSessionID, userID string
	var refreshExpires time.Time
	if err := tx.QueryRowContext(ctx, `SELECT id, user_id, refresh_expires_at FROM sessions WHERE refresh_token_hash = $1 AND revoked_at IS NULL FOR UPDATE`, HashToken(refreshToken)).Scan(&oldSessionID, &userID, &refreshExpires); err != nil || !refreshExpires.After(now) {
		return User{}, Session{}, ErrInvalidCredentials
	}
	var user User
	if err := tx.QueryRowContext(ctx, `SELECT id, openim_user_id, username, email, display_name, avatar_url, status, role FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(&user.ID, &user.OpenIMUserID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role); err != nil || user.Status != "active" {
		return User{}, Session{}, ErrInvalidCredentials
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at = $1, last_seen_at = $1 WHERE id = $2`, now, oldSessionID); err != nil {
		return User{}, Session{}, fmt.Errorf("revoke old session: %w", err)
	}
	session, err := newSession(user.ID, now, s.AccessTokenTTL, s.RefreshTokenTTL)
	if err != nil {
		return User{}, Session{}, err
	}
	if err := insertSession(ctx, tx, user.ID, session, userAgent, ip, now); err != nil {
		return User{}, Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, Session{}, fmt.Errorf("commit refresh: %w", err)
	}
	return user, session, nil
}

func (s Service) Logout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil
	}
	_, err := s.DB.ExecContext(ctx, `UPDATE sessions SET revoked_at = COALESCE(revoked_at, $1), last_seen_at = $1 WHERE access_token_hash = $2`, time.Now().UTC(), HashToken(accessToken))
	return err
}

func (s Service) UserByAccessToken(ctx context.Context, accessToken string) (User, error) {
	var user User
	err := s.DB.QueryRowContext(ctx, `SELECT u.id, u.openim_user_id, u.username, u.email, u.display_name, u.avatar_url, u.status, u.role FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.access_token_hash = $1 AND s.revoked_at IS NULL AND s.access_expires_at > now() AND u.status = 'active' AND u.deleted_at IS NULL`, HashToken(accessToken)).Scan(&user.ID, &user.OpenIMUserID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

type fieldValidationError struct {
	Field string
	Err   error
}

func (e fieldValidationError) Error() string   { return e.Err.Error() }
func (e fieldValidationError) Unwrap() error   { return e.Err }
func fieldError(field string, err error) error { return fieldValidationError{Field: field, Err: err} }

func FieldName(err error) string {
	var fieldErr fieldValidationError
	if errors.As(err, &fieldErr) {
		return fieldErr.Field
	}
	return ""
}

func newSession(userID string, now time.Time, accessTTL, refreshTTL time.Duration) (Session, error) {
	access, err := NewOpaqueToken()
	if err != nil {
		return Session{}, err
	}
	refresh, err := NewOpaqueToken()
	if err != nil {
		return Session{}, err
	}
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	id, err := NewUUID()
	if err != nil {
		return Session{}, err
	}
	return Session{ID: id, AccessToken: access, AccessExpiresAt: now.Add(accessTTL), RefreshToken: refresh, RefreshExpiresAt: now.Add(refreshTTL)}, nil
}

func insertSession(ctx context.Context, tx *sql.Tx, userID string, session Session, userAgent string, ip net.IP, now time.Time) error {
	var ipValue any
	if ip != nil {
		ipValue = ip.String()
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO sessions (id, user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at, user_agent, ip_address, last_seen_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`, session.ID, userID, HashToken(session.AccessToken), HashToken(session.RefreshToken), session.AccessExpiresAt, session.RefreshExpiresAt, userAgent, ipValue, now)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func normalizeDBError(err error) error {
	if strings.Contains(err.Error(), "duplicate key") {
		return errors.New("username or email is already registered")
	}
	return fmt.Errorf("database operation failed: %w", err)
}
