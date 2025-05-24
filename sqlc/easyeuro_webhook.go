package db

import (
	"context"
	"strings"
	"time"
)

type EasyeuroWebhook struct {
	ID                 string    `json:"id"`
	URL                string    `json:"url"`
	EventTypes         []string  `json:"eventTypes"`
	SignatureAlgorithm string    `json:"signatureAlgorithm"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (store *SQLStore) EasyeuroWebhookFindOne(ctx context.Context, id string) (*EasyeuroWebhook, error) {
	query := `SELECT * FROM easyeuro_webhook WHERE id = $1`
	var webhook EasyeuroWebhook
	var eventTypes string
	err := store.db.QueryRowContext(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.URL,
		&eventTypes,
		&webhook.SignatureAlgorithm,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	webhook.EventTypes = strings.Split(eventTypes, ",")
	return &webhook, nil
}

func (store *SQLStore) EasyeuroWebhookFindOneByURLAndEventType(ctx context.Context, url string, eventType string) (*EasyeuroWebhook, error) {
	query := `SELECT * FROM easyeuro_webhook WHERE url = $1 AND event_types = $2`

	var webhook EasyeuroWebhook
	var eventTypes string
	err := store.db.QueryRowContext(ctx, query, url, eventType).Scan(
		&webhook.ID,
		&webhook.URL,
		&eventTypes,
		&webhook.SignatureAlgorithm,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	webhook.EventTypes = strings.Split(eventTypes, ",")
	return &webhook, nil
}

func (store *SQLStore) EasyeuroWebhookCreate(ctx context.Context, arg EasyeuroWebhook) (*EasyeuroWebhook, error) {
	query := `INSERT INTO easyeuro_webhook (id, url, event_types, signature_algorithm) VALUES ($1, $2, $3, $4) RETURNING *`

	eventTypes := strings.Join(arg.EventTypes, ",")

	err := store.db.QueryRowContext(ctx, query, arg.ID, arg.URL, eventTypes, arg.SignatureAlgorithm).Scan(
		&arg.ID,
		&arg.URL,
		&eventTypes,
		&arg.SignatureAlgorithm,
		&arg.CreatedAt,
		&arg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	arg.EventTypes = strings.Split(eventTypes, ",")
	return &arg, nil
}
