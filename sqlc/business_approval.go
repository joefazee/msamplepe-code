package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type BusinessStatus string

const (
	BusinessStatusPending   BusinessStatus = "pending"
	BusinessStatusNew       BusinessStatus = "new"
	BusinessStatusSubmitted BusinessStatus = "submitted"
	BusinessStatusFailed    BusinessStatus = "submitted"
	BusinessStatusPartial   BusinessStatus = "partial-submitted"
	BusinessStatusApproved  BusinessStatus = "approved"
	BusinessStatusRejected  BusinessStatus = "rejected"
)

func (bs BusinessStatus) String() string {
	return string(bs)
}

// updateApprovalStatus updates the approval fields for a given table.
func (q *Queries) updateApprovalStatus(ctx context.Context, table string, id uuid.UUID, status BusinessStatus, reason string, approvedBy uuid.UUID) error {
	query := fmt.Sprintf(
		`UPDATE %s SET approval_status = $1, approval_status_reason = $2, approval_status_updated_by = $3, approval_status_updated_at = NOW() WHERE id = $4`, table)
	_, err := q.db.ExecContext(ctx, query, status.String(), reason, approvedBy, id)
	return err
}

// SetBusinessApprovalStatus updates the approval status in the "businesses" table.
func (q *Queries) SetBusinessApprovalStatus(ctx context.Context, id uuid.UUID, status BusinessStatus, reason string, approvedBy uuid.UUID) error {
	return q.updateApprovalStatus(ctx, "businesses", id, status, reason, approvedBy)
}

// SetBusinessOwnersApprovalStatus updates the approval status in the "business_owners" table.
func (q *Queries) SetBusinessOwnersApprovalStatus(ctx context.Context, id uuid.UUID, status BusinessStatus, reason string, approvedBy uuid.UUID) error {
	return q.updateApprovalStatus(ctx, "business_owners", id, status, reason, approvedBy)
}
