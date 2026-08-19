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

type Service struct {
	DB              *sql.DB
	BcryptCost      int
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type User struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"-"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Status      string  `json:"-"`
	Role        string  `json:"-"`
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
	cost := s.BcryptCost
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), cost)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	userID, err := NewUUID()
	if err != nil {
		return User{}, Session{}, err
	}
	user := User{ID: userID, Username: username, Email: email, DisplayName: displayName, Status: "active", Role: "user"}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("begin register: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO users (id, username, email, password_hash, display_name, status, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, 'active', 'user', $6, $6)`, user.ID, user.Username, user.Email, string(passwordHash), user.DisplayName, now)
	if err != nil {
		return User{}, Session{}, normalizeDBError(err)
	}
	session, err := newSession(user.ID, userAgent, ip, now, s.AccessTokenTTL, s.RefreshTokenTTL)
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

func (s Service) Login(ctx context.Context, input LoginInput, userAgent string, ip net.IP) (User, Session, error) {
	login := strings.ToLower(strings.TrimSpace(input.Login))
	var user User
	var passwordHash string
	err := s.DB.QueryRowContext(ctx, `SELECT id, username, email, password_hash, display_name, avatar_url, status, role FROM users WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`, login).Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role)
	if err != nil {
		return User{}, Session{}, ErrInvalidCredentials
	}
	if user.Status != "active" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		return User{}, Session{}, ErrInvalidCredentials
	}
	now := time.Now().UTC()
	session, err := newSession(user.ID, userAgent, ip, now, s.AccessTokenTTL, s.RefreshTokenTTL)
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
	if err := tx.QueryRowContext(ctx, `SELECT id, username, email, display_name, avatar_url, status, role FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role); err != nil || user.Status != "active" {
		return User{}, Session{}, ErrInvalidCredentials
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at = $1, last_seen_at = $1 WHERE id = $2`, now, oldSessionID); err != nil {
		return User{}, Session{}, fmt.Errorf("revoke old session: %w", err)
	}
	session, err := newSession(user.ID, userAgent, ip, now, s.AccessTokenTTL, s.RefreshTokenTTL)
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
	err := s.DB.QueryRowContext(ctx, `SELECT u.id, u.username, u.email, u.display_name, u.avatar_url, u.status, u.role FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.access_token_hash = $1 AND s.revoked_at IS NULL AND s.access_expires_at > now() AND u.status = 'active' AND u.deleted_at IS NULL`, HashToken(accessToken)).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s Service) AuthenticateTinode(ctx context.Context, username, password string) (User, string, error) {
	var user User
	var passwordHash string
	var rawUID, publicUID sql.NullString
	err := s.DB.QueryRowContext(ctx, `SELECT id, username, email, password_hash, display_name, avatar_url, status, role, tinode_uid_raw, tinode_user_id FROM users WHERE username = $1 AND deleted_at IS NULL`, username).Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.DisplayName, &user.AvatarURL, &user.Status, &user.Role, &rawUID, &publicUID)
	if err != nil || user.Status != "active" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return User{}, "", ErrInvalidCredentials
	}
	if rawUID.Valid {
		return user, rawUID.String, nil
	}
	return user, "", nil
}

func (s Service) LinkTinode(ctx context.Context, username, uid string) error {
	if uid == "" {
		return fmt.Errorf("uid is required")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Tinode link: %w", err)
	}
	defer tx.Rollback()
	var current sql.NullString
	var userID string
	if err := tx.QueryRowContext(ctx, `SELECT id, tinode_uid_raw FROM users WHERE username = $1 AND deleted_at IS NULL FOR UPDATE`, username).Scan(&userID, &current); err != nil {
		return ErrInvalidCredentials
	}
	if current.Valid && current.String != uid {
		return fmt.Errorf("tinode uid conflict")
	}
	var owner string
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE tinode_uid_raw = $1`, uid).Scan(&owner)
	if err == nil && owner != userID {
		return fmt.Errorf("tinode uid conflict")
	}
	if !current.Valid {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET tinode_uid_raw = $1, tinode_user_id = $1, updated_at = $2 WHERE id = $3`, uid, time.Now().UTC(), userID); err != nil {
			return fmt.Errorf("store Tinode link: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Tinode link: %w", err)
	}
	return nil
}

func (s Service) TinodeUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`, username).Scan(&exists); err != nil {
		return false, err
	}
	return !exists, nil
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

func newSession(userID string, userAgent string, ip net.IP, now time.Time, accessTTL, refreshTTL time.Duration) (Session, error) {
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
		return fmt.Errorf("username or email is already registered")
	}
	return fmt.Errorf("database operation failed: %w", err)
}
