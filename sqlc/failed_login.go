package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type FailedLogin struct {
	ID             int
	UserIdentity   string
	Provider       string
	FailedAttempts int
	LastFailedAt   time.Time
	BannedUntil    sql.NullTime
}

func (store *SQLStore) GetFailedLogin(ctx context.Context, userIdentity, provider string) (*FailedLogin, error) {
	query := `
        SELECT id, user_identity, provider, failed_attempts, last_failed_at, banned_until
        FROM user_failed_logins
        WHERE user_identity = $1 AND provider = $2
    `
	row := store.db.QueryRowContext(ctx, query, userIdentity, provider)

	var failedLogin FailedLogin
	err := row.Scan(&failedLogin.ID, &failedLogin.UserIdentity, &failedLogin.Provider, &failedLogin.FailedAttempts, &failedLogin.LastFailedAt, &failedLogin.BannedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch login attempts: %w", err)
	}
	return &failedLogin, nil
}

func (store *SQLStore) CreateFailedLogin(ctx context.Context, userIdentity, provider string) error {
	query := `
        INSERT INTO user_failed_logins (user_identity, provider, failed_attempts, last_failed_at)
        VALUES ($1, $2, 1, CURRENT_TIMESTAMP)
        ON CONFLICT (user_identity, provider) DO NOTHING
    `
	_, err := store.db.ExecContext(ctx, query, userIdentity, provider)
	if err != nil {
		return fmt.Errorf("failed to create failed login record: %w", err)
	}
	return nil
}

func (store *SQLStore) UpdateFailedLogin(ctx context.Context, failedLogin *FailedLogin) error {
	log.Printf("failedLogin: %+v", failedLogin)
	query := `
        UPDATE user_failed_logins
        SET failed_attempts = $1, last_failed_at = CURRENT_TIMESTAMP, banned_until = $2
        WHERE user_identity = $3 AND provider = $4
    `
	_, err := store.db.ExecContext(ctx, query, failedLogin.FailedAttempts, failedLogin.BannedUntil.Time, failedLogin.UserIdentity, failedLogin.Provider)
	if err != nil {
		return fmt.Errorf("failed to update failed login record: %w", err)
	}
	return nil
}

func (store *SQLStore) ResetFailedLogin(ctx context.Context, userIdentity, provider string) error {
	query := `
        UPDATE user_failed_logins
        SET failed_attempts = 0, banned_until = NULL
        WHERE user_identity = $1 AND provider = $2
    `
	_, err := store.db.ExecContext(ctx, query, userIdentity, provider)
	if err != nil {
		return fmt.Errorf("failed to reset failed login record: %w", err)
	}
	return nil
}

func (store *SQLStore) GetBannedUntil(ctx context.Context, userIdentity, provider string) (*time.Time, error) {
	query := `
        SELECT banned_until
        FROM user_failed_logins
        WHERE user_identity = $1 AND provider = $2
    `
	var bannedUntil sql.NullTime
	err := store.db.QueryRowContext(ctx, query, userIdentity, provider).Scan(&bannedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch banned_until: %w", err)
	}

	if bannedUntil.Valid {
		return &bannedUntil.Time, nil
	}
	return nil, nil
}
