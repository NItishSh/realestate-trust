package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// OutboxRecord represents an entry in the outbox_events table.
type OutboxRecord struct {
	ID            string    `json:"id"`
	AggregateType string    `json:"aggregateType"`
	AggregateID   string    `json:"aggregateId"`
	EventType     string    `json:"eventType"`
	Payload       string    `json:"payload"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// OutboxRelay is a background worker that polls and relays pending outbox events to RabbitMQ.
type OutboxRelay struct {
	db        *sql.DB
	publisher *Publisher
	queueName string
	interval  time.Duration
	stopChan  chan struct{}
}

// NewOutboxRelay creates a new outbox relay worker.
func NewOutboxRelay(db *sql.DB, conn *amqp.Connection, queueName string, interval time.Duration) (*OutboxRelay, error) {
	pub, err := NewPublisher(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher for outbox relay: %w", err)
	}

	if interval <= 0 {
		interval = 500 * time.Millisecond
	}

	return &OutboxRelay{
		db:        db,
		publisher: pub,
		queueName: queueName,
		interval:  interval,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start runs the outbox relay loop in a background goroutine.
func (r *OutboxRelay) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		defer func() { _ = r.publisher.Close() }()

		slog.Info("Transactional Outbox Relay worker started", "queue", r.queueName, "poll_interval", r.interval)
		for {
			select {
			case <-r.stopChan:
				slog.Info("Transactional Outbox Relay worker stopped")
				return
			case <-ctx.Done():
				slog.Info("Context cancelled; stopping Outbox Relay worker")
				return
			case <-ticker.C:
				if err := r.ProcessBatch(ctx); err != nil {
					slog.ErrorContext(ctx, "Outbox relay batch processing error", "err", err)
				}
			}
		}
	}()
}

// Stop stops the outbox relay worker.
func (r *OutboxRelay) Stop() {
	close(r.stopChan)
}

// ProcessBatch polls and publishes a batch of pending outbox events using SKIP LOCKED.
func (r *OutboxRelay) ProcessBatch(ctx context.Context) error {
	if r.db == nil {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `SELECT id, aggregate_type, aggregate_id, event_type, payload
	          FROM outbox_events
	          WHERE status = 'PENDING'
	          ORDER BY created_at ASC
	          LIMIT 50
	          FOR UPDATE SKIP LOCKED`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	type eventItem struct {
		id            string
		aggregateType string
		aggregateID   string
		eventType     string
		payload       string
	}

	var items []eventItem
	for rows.Next() {
		var it eventItem
		if err := rows.Scan(&it.id, &it.aggregateType, &it.aggregateID, &it.eventType, &it.payload); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, it)
	}
	_ = rows.Close()

	if len(items) == 0 {
		return nil
	}

	publishedIDs := make([]string, 0, len(items))
	for _, it := range items {
		// Convert CloudEvent or generic payload to TransactionEvent for the queue
		var event TransactionEvent
		if err := json.Unmarshal([]byte(it.payload), &event); err != nil || event.ID == "" {
			// Wrap in TransactionEvent for backward-compatible consumer parsing
			event = TransactionEvent{
				ID:        it.id,
				Action:    it.eventType,
				Payload:   it.payload,
				Timestamp: time.Now().UTC(),
			}
		}

		body, err := json.Marshal(event)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal outbox event for publishing", "id", it.id, "err", err)
			continue
		}

		if err := r.publisher.Publish(ctx, r.queueName, body, it.id); err != nil {
			slog.WarnContext(ctx, "Failed to publish outbox event to RabbitMQ, will retry on next poll", "id", it.id, "err", err)
			// Increment retry_count on the outbox record
			_, _ = tx.ExecContext(ctx, `UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id = $1`, it.id)
			continue
		}

		publishedIDs = append(publishedIDs, it.id)
	}

	if len(publishedIDs) > 0 {
		for _, id := range publishedIDs {
			_, err := tx.ExecContext(ctx, `UPDATE outbox_events SET status = 'PUBLISHED', published_at = NOW() WHERE id = $1`, id)
			if err != nil {
				return err
			}
		}
		slog.InfoContext(ctx, "Successfully relayed outbox events to RabbitMQ", "count", len(publishedIDs))
	}

	return tx.Commit()
}
