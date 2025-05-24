package db

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

const (
	DatatypeString = "string"

	DatatypeBoolean = "bool"

	IdentityVerifiedTypeVeriff = "veriff"
	IdentityVerifiedTypeAuto   = "auto"

	IdentityVerificationStatusPending   = "pending"
	IdentityVerificationStatusFailed    = "failed"
	IdentityVerificationStatusNew       = "new"
	IdentityVerificationStatusApproved  = "approved"
	IdentityVerificationStatusSubmitted = "submitted"
)

type UserMeta struct {
	IdentityVerified           bool   `json:"identity_verified"`
	IdentityVerificationType   string `json:"identity_verification_type"`
	IdentityVerificationDoc    string `json:"identity_verification_doc"`
	IdentityVerificationStatus string `json:"identity_verification_status"`
}

func (q *Queries) GetUserMetas(ctx context.Context, userID uuid.UUID) (UserMeta, error) {
	var usemeta UserMeta

	rows, err := q.db.QueryContext(ctx, "SELECT key, value, datatype FROM user_meta WHERE user_id = $1", userID)
	if err != nil {
		return usemeta, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value, datatype string
		if err := rows.Scan(
			&key, &value, &datatype,
		); err != nil {
			return usemeta, err
		}
		mapUserMeta(&usemeta, key, value, datatype)
	}
	if err := rows.Close(); err != nil {
		return usemeta, err
	}
	if err := rows.Err(); err != nil {
		return usemeta, err
	}

	return usemeta, nil
}

type UserMetaCreateParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Key      string    `json:"key"`
	Value    string    `json:"value"`
	Datatype string    `json:"datatype"`
}

func (q *Queries) UserMetaExists(ctx context.Context, userID uuid.UUID, key string) (bool, error) {
	var exists bool
	err := q.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM user_meta WHERE user_id = $1 AND key = $2)", userID, key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (q *Queries) SetUserMeta(ctx context.Context, input UserMetaCreateParams) error {

	exists, err := q.UserMetaExists(ctx, input.UserID, input.Key)
	if err != nil {
		return err
	}

	if exists {
		_, err := q.db.ExecContext(ctx, "UPDATE user_meta SET value = $1 WHERE user_id = $2 AND key = $3", input.Value, input.UserID, input.Key)
		return err
	}
	_, err = q.db.ExecContext(ctx, "INSERT INTO user_meta (user_id, key, value, datatype) VALUES ($1, $2, $3, $4)", input.UserID, input.Key, input.Value, input.Datatype)
	return err
}

func mapUserMeta(u *UserMeta, key string, value string, datatype string) {
	switch datatype {
	case DatatypeBoolean:
		if key == "identity_verified" {
			u.IdentityVerified = strings.ToLower(value) == "true"
		}
	case DatatypeString:
		if key == "identity_verification_type" {
			u.IdentityVerificationType = value
		}

		if key == "identity_verification_doc" {
			u.IdentityVerificationDoc = value
		}

		if key == "identity_verification_status" {
			u.IdentityVerificationStatus = value
		}
	}

}
