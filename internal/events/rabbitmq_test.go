package events

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestRabbitMQ_CorrelationIDPropagation(t *testing.T) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		t.Skip("RABBITMQ_URL is not set; skipping live integration tests")
	}

	conn, err := Connect(rabbitURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	const queueName = "test-correlation-queue"
	const expectedCorrelationID = "test-correlation-12345"

	// Create context with correlation_id
	ctx := context.WithValue(context.Background(), CorrelationIDContextKey, expectedCorrelationID)

	event := TransactionEvent{
		ID:        "evt-test-123",
		Action:    "test",
		Payload:   "correlation test payload",
		Timestamp: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedCorrelationID string
	var receivedEvent TransactionEvent

	err = Consume(conn, queueName, func(c context.Context, e TransactionEvent) error {
		defer wg.Done()
		receivedEvent = e
		if c != nil {
			if cid, ok := c.Value("correlation_id").(string); ok {
				receivedCorrelationID = cid
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to start consumer: %v", err)
	}

	// Publish with correlation ID context
	err = Publish(ctx, conn, queueName, event)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// Wait for consumer (timeout 5s)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for consumer to process event")
	}

	if receivedEvent.ID != event.ID {
		t.Errorf("expected event ID %s, got %s", event.ID, receivedEvent.ID)
	}

	if receivedCorrelationID != expectedCorrelationID {
		t.Errorf("expected correlation ID %s, got %s", expectedCorrelationID, receivedCorrelationID)
	}
}
