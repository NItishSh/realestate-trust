package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestCloudEvent_Serialization(t *testing.T) {
	createdEvent := CloudEvent[TransactionCreatedPayload]{
		SpecVersion: "1.0",
		ID:          "event-uuid-123",
		Source:      "realestate-trust/transaction-manager",
		Type:        "com.realestatetrust.transaction.created.v1",
		Time:        time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
		Data: TransactionCreatedPayload{
			TransactionID: "tx-prop-001",
			PropertyID:    "prop-001",
			BuyerID:       "user-buyer-001",
			SellerID:      "user-seller-001",
			Amount:        5000000.0,
			Currency:      "INR",
		},
	}

	bytes, err := json.Marshal(createdEvent)
	if err != nil {
		t.Fatalf("failed to marshal CloudEvent: %v", err)
	}

	var deserialized CloudEvent[TransactionCreatedPayload]
	if err := json.Unmarshal(bytes, &deserialized); err != nil {
		t.Fatalf("failed to unmarshal CloudEvent: %v", err)
	}

	if deserialized.ID != "event-uuid-123" {
		t.Errorf("expected ID event-uuid-123, got %s", deserialized.ID)
	}
	if deserialized.Data.TransactionID != "tx-prop-001" {
		t.Errorf("expected transaction ID tx-prop-001, got %s", deserialized.Data.TransactionID)
	}
	if deserialized.Data.Amount != 5000000.0 {
		t.Errorf("expected amount 5000000.0, got %f", deserialized.Data.Amount)
	}
}

func TestTransactionStatusUpdatedPayload_Serialization(t *testing.T) {
	updatedEvent := CloudEvent[TransactionStatusUpdatedPayload]{
		SpecVersion: "1.0",
		ID:          "event-uuid-456",
		Source:      "realestate-trust/transaction-manager",
		Type:        "com.realestatetrust.transaction.updated.v1",
		Time:        time.Now().UTC(),
		Data: TransactionStatusUpdatedPayload{
			TransactionID: "tx-prop-001",
			NewState:      "FUNDED",
		},
	}

	bytes, err := json.Marshal(updatedEvent)
	if err != nil {
		t.Fatalf("failed to marshal CloudEvent: %v", err)
	}

	var deserialized CloudEvent[TransactionStatusUpdatedPayload]
	if err := json.Unmarshal(bytes, &deserialized); err != nil {
		t.Fatalf("failed to unmarshal CloudEvent: %v", err)
	}

	if deserialized.Data.NewState != "FUNDED" {
		t.Errorf("expected new state FUNDED, got %s", deserialized.Data.NewState)
	}
}

func TestOutboxRelay_LiveBatchProcessing(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if dbURL == "" || rabbitURL == "" {
		t.Skip("DATABASE_URL or RABBITMQ_URL not set; skipping live outbox relay test")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer func() { _ = db.Close() }()

	conn, err := Connect(rabbitURL)
	if err != nil {
		t.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Clean outbox table
	_, _ = db.Exec("DELETE FROM outbox_events")

	// Insert a pending outbox event
	eventID := "outbox-test-live-01"
	payload := `{"id":"outbox-test-live-01","action":"created","payload":"Test Outbox Relay Event","timestamp":"2026-09-01T12:00:00Z"}`
	insertQuery := `INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status)
	                VALUES ($1, 'transaction', 'tx-test-01', 'com.realestatetrust.transaction.created.v1', $2, 'PENDING')`
	_, err = db.Exec(insertQuery, eventID, payload)
	if err != nil {
		t.Fatalf("failed to insert outbox event: %v", err)
	}

	queueName := "test-outbox-events-queue"
	relay, err := NewOutboxRelay(db, conn, queueName, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create OutboxRelay: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Process batch
	err = relay.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("failed to process outbox batch: %v", err)
	}

	// Verify status changed to PUBLISHED
	var status string
	var publishedAt sql.NullTime
	err = db.QueryRow("SELECT status, published_at FROM outbox_events WHERE id = $1", eventID).Scan(&status, &publishedAt)
	if err != nil {
		t.Fatalf("failed to query outbox event status: %v", err)
	}

	if status != "PUBLISHED" {
		t.Errorf("expected outbox event status PUBLISHED, got %s", status)
	}
	if !publishedAt.Valid {
		t.Errorf("expected published_at to be valid timestamp")
	}
}
