// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: affiliate_merchant_webhook_tmp.sql

package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const createAffiliateMerchantWebhookMetaData = `-- name: CreateAffiliateMerchantWebhookMetaData :one
INSERT INTO affiliate_merchant_webhook_tmp (merchant_id, swap_data, meta_data, webhook_status)
    VALUES ($1, $2, $3, $4) RETURNING id, merchant_id, swap_data, webhook_status, meta_data, created_at, updated_at
`

type CreateAffiliateMerchantWebhookMetaDataParams struct {
	MerchantID    uuid.UUID       `json:"merchant_id"`
	SwapData      json.RawMessage `json:"swap_data"`
	MetaData      json.RawMessage `json:"meta_data"`
	WebhookStatus int32           `json:"webhook_status"`
}

func (q *Queries) CreateAffiliateMerchantWebhookMetaData(ctx context.Context, arg CreateAffiliateMerchantWebhookMetaDataParams) (AffiliateMerchantWebhookTmp, error) {
	row := q.db.QueryRowContext(ctx, createAffiliateMerchantWebhookMetaData,
		arg.MerchantID,
		arg.SwapData,
		arg.MetaData,
		arg.WebhookStatus,
	)
	var i AffiliateMerchantWebhookTmp
	err := row.Scan(
		&i.ID,
		&i.MerchantID,
		&i.SwapData,
		&i.WebhookStatus,
		&i.MetaData,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAffiliateMerchantWebhookMetaDatasByWebhookStatus = `-- name: GetAffiliateMerchantWebhookMetaDatasByWebhookStatus :many
SELECT id, merchant_id, swap_data, webhook_status, meta_data, created_at, updated_at FROM affiliate_merchant_webhook_tmp WHERE webhook_status = $1 LIMIT 50
`

func (q *Queries) GetAffiliateMerchantWebhookMetaDatasByWebhookStatus(ctx context.Context, webhookStatus int32) ([]AffiliateMerchantWebhookTmp, error) {
	rows, err := q.db.QueryContext(ctx, getAffiliateMerchantWebhookMetaDatasByWebhookStatus, webhookStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AffiliateMerchantWebhookTmp{}
	for rows.Next() {
		var i AffiliateMerchantWebhookTmp
		if err := rows.Scan(
			&i.ID,
			&i.MerchantID,
			&i.SwapData,
			&i.WebhookStatus,
			&i.MetaData,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateAffiliateMerchantWebhookStatus = `-- name: UpdateAffiliateMerchantWebhookStatus :one
UPDATE affiliate_merchant_webhook_tmp SET webhook_status = $1 WHERE id = $2 RETURNING id, merchant_id, swap_data, webhook_status, meta_data, created_at, updated_at
`

type UpdateAffiliateMerchantWebhookStatusParams struct {
	WebhookStatus int32     `json:"webhook_status"`
	ID            uuid.UUID `json:"id"`
}

func (q *Queries) UpdateAffiliateMerchantWebhookStatus(ctx context.Context, arg UpdateAffiliateMerchantWebhookStatusParams) (AffiliateMerchantWebhookTmp, error) {
	row := q.db.QueryRowContext(ctx, updateAffiliateMerchantWebhookStatus, arg.WebhookStatus, arg.ID)
	var i AffiliateMerchantWebhookTmp
	err := row.Scan(
		&i.ID,
		&i.MerchantID,
		&i.SwapData,
		&i.WebhookStatus,
		&i.MetaData,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
