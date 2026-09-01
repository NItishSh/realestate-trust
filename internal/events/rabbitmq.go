package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type contextKey string

const CorrelationIDContextKey contextKey = "correlation_id"

// TransactionEvent represents a legacy/generic event payload for backward compatibility.
type TransactionEvent struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// CloudEvent represents a CNCF CloudEvents v1.0 compliant envelope.
type CloudEvent[T any] struct {
	SpecVersion     string    `json:"specversion"`
	ID              string    `json:"id"`
	Source          string    `json:"source"`
	Type            string    `json:"type"`
	Time            time.Time `json:"time"`
	DataContentType string    `json:"datacontenttype,omitempty"`
	Data            T         `json:"data"`
}

// Domain Event Payload DTOs
type TransactionCreatedPayload struct {
	TransactionID string  `json:"transactionId"`
	PropertyID    string  `json:"propertyId"`
	BuyerID       string  `json:"buyerId"`
	SellerID      string  `json:"sellerId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

type TransactionStatusUpdatedPayload struct {
	TransactionID string `json:"transactionId"`
	NewState      string `json:"newState"`
}

type EscrowFundedPayload struct {
	TransactionID string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

// Connect establishes a connection to RabbitMQ
func Connect(url string) (*amqp.Connection, error) {
	if url == "" {
		return nil, fmt.Errorf("rabbitmq url cannot be empty")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}
	return conn, nil
}

// setupQueueTopology sets up the main queue, retry queue (with TTL), and dead-letter exchange/queue.
func setupQueueTopology(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	dlxName := queueName + "-dlx"
	dlqName := queueName + "-dlq"
	retryQueueName := queueName + "-retry"

	// 1. Declare DLX Exchange
	err := ch.ExchangeDeclare(
		dlxName,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// 2. Declare DLQ Queue
	_, err = ch.QueueDeclare(
		dlqName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// 3. Bind DLQ to DLX
	err = ch.QueueBind(
		dlqName,
		"dead-letter",
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}

	// 4. Declare Retry Queue with TTL delay back to main queue
	retryArgs := amqp.Table{
		"x-message-ttl":             int32(3000), // 3-second retry delay
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": queueName, // route back to main queue on TTL expire
	}
	_, err = ch.QueueDeclare(
		retryQueueName,
		true,
		false,
		false,
		false,
		retryArgs,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare retry queue: %w", err)
	}

	// 5. Declare Main Queue with DLX
	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": "dead-letter",
	}

	return ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		mainArgs,
	)
}

// Publisher maintains a thread-safe, long-lived AMQP channel for high-throughput message publishing.
type Publisher struct {
	conn           *amqp.Connection
	ch             *amqp.Channel
	declaredQueues map[string]bool
	mu             sync.Mutex
}

// NewPublisher creates a new long-lived channel publisher.
func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open publisher channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	return &Publisher{
		conn:           conn,
		ch:             ch,
		declaredQueues: make(map[string]bool),
	}, nil
}

// Close closes the underlying channel.
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		return p.ch.Close()
	}
	return nil
}

// Publish sends a raw JSON message to the specified queue with publisher confirms.
func (p *Publisher) Publish(ctx context.Context, queueName string, body []byte, correlationID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Ensure topology is declared once per queue
	if !p.declaredQueues[queueName] {
		if _, err := setupQueueTopology(p.ch, queueName); err != nil {
			return err
		}
		p.declaredQueues[queueName] = true
	}

	confirmChan := p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.ch.PublishWithContext(pubCtx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"correlation_id": correlationID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	select {
	case confirm := <-confirmChan:
		if !confirm.Ack {
			return fmt.Errorf("message not acknowledged by RabbitMQ broker")
		}
	case <-pubCtx.Done():
		return fmt.Errorf("publish confirmation timed out")
	}

	return nil
}

// Publish is a convenience function that sends a TransactionEvent using a channel.
func Publish(ctx context.Context, conn *amqp.Connection, queueName string, event TransactionEvent) error {
	pub, err := NewPublisher(conn)
	if err != nil {
		return err
	}
	defer func() { _ = pub.Close() }()

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	cid := ""
	if ctx != nil {
		if val, ok := ctx.Value(CorrelationIDContextKey).(string); ok {
			cid = val
		} else if val, ok := ctx.Value("correlation_id").(string); ok {
			cid = val
		}
	}

	return pub.Publish(ctx, queueName, body, cid)
}

// Consume runs a background loop with retry backoff before dead-lettering.
func Consume(conn *amqp.Connection, queueName string, handler func(context.Context, TransactionEvent) error) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := setupQueueTopology(ch, queueName)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to setup queue topology: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // manual acks
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		defer func() { _ = ch.Close() }()
		slog.Info("Started consuming events from RabbitMQ", "queue", queueName)
		for d := range msgs {
			var event TransactionEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				slog.Error("failed to unmarshal message payload", "err", err)
				// Poison message: route directly to DLQ
				_ = d.Nack(false, false)
				continue
			}

			cid := ""
			if val, ok := d.Headers["correlation_id"].(string); ok {
				cid = val
			} else if valBytes, ok := d.Headers["correlation_id"].([]byte); ok {
				cid = string(valBytes)
			}

			// Check retry count
			retryCount := int64(0)
			if rVal, ok := d.Headers["x-retry-count"].(int64); ok {
				retryCount = rVal
			} else if rValInt, ok := d.Headers["x-retry-count"].(int32); ok {
				retryCount = int64(rValInt)
			}

			ctx := context.WithValue(context.Background(), CorrelationIDContextKey, cid)

			slog.InfoContext(ctx, "Event received from RabbitMQ", "queue", queueName, "action", event.Action, "id", event.ID, "retry", retryCount)
			if err := handler(ctx, event); err != nil {
				slog.ErrorContext(ctx, "failed to handle message", "err", err, "retry", retryCount)

				if retryCount < 3 {
					// Route to retry queue with delay
					retryHeaders := amqp.Table{
						"correlation_id": cid,
						"x-retry-count":  retryCount + 1,
					}
					_ = ch.PublishWithContext(context.Background(),
						"",
						queueName+"-retry",
						false,
						false,
						amqp.Publishing{
							ContentType: "application/json",
							Body:        d.Body,
							Headers:     retryHeaders,
						},
					)
					_ = d.Ack(false) // Ack from main queue since it is now in retry queue
				} else {
					slog.WarnContext(ctx, "Max retries exceeded; routing message to DLQ", "id", event.ID)
					_ = d.Nack(false, false) // Route to DLQ
				}
			} else {
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}
