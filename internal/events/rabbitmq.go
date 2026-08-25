package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type contextKey string

const CorrelationIDContextKey contextKey = "correlation_id"

type TransactionEvent struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
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

// setupQueue declares a Dead Letter Exchange (DLX), a Dead Letter Queue (DLQ),
// binds them, and then declares the main queue configured with the DLX.
func setupQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	dlxName := queueName + "-dlx"
	dlqName := queueName + "-dlq"

	// 1. Declare DLX Exchange
	err := ch.ExchangeDeclare(
		dlxName,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // arguments
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
		nil,   // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// 3. Bind DLQ to DLX
	err = ch.QueueBind(
		dlqName,
		"dead-letter", // routing key
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}

	// 4. Declare main queue with DLX args
	args := amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": "dead-letter",
	}

	return ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
}

// Publish sends a message to the specified queue with Publisher Confirms enabled
func Publish(ctx context.Context, conn *amqp.Connection, queueName string, event TransactionEvent) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() { _ = ch.Close() }()

	// Setup queue with DLX/DLQ
	q, err := setupQueue(ch, queueName)
	if err != nil {
		return fmt.Errorf("failed to setup queue: %w", err)
	}

	// Enable Publisher Confirms
	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	confirmChan := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cid := ""
	if ctx != nil {
		if val, ok := ctx.Value(CorrelationIDContextKey).(string); ok {
			cid = val
		}
	}

	err = ch.PublishWithContext(pubCtx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"correlation_id": cid,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for publisher confirmation
	select {
	case confirm := <-confirmChan:
		if !confirm.Ack {
			return fmt.Errorf("message not acknowledged by RabbitMQ broker")
		}
	case <-pubCtx.Done():
		return fmt.Errorf("publish confirmation timed out")
	}

	slog.InfoContext(ctx, "Event published to RabbitMQ with confirmation", "queue", queueName, "action", event.Action, "id", event.ID)
	return nil
}

// Consume runs a background loop to process messages from a queue with manual acknowledgements
func Consume(conn *amqp.Connection, queueName string, handler func(context.Context, TransactionEvent) error) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Setup queue with DLX/DLQ
	q, err := setupQueue(ch, queueName)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("failed to setup queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (manual confirmation enabled)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
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
				// Negative acknowledge without requeuing (routes to DLQ)
				_ = d.Nack(false, false)
				continue
			}

			// Extract correlation_id from message headers
			cid := ""
			if val, ok := d.Headers["correlation_id"].(string); ok {
				cid = val
			} else if valBytes, ok := d.Headers["correlation_id"].([]byte); ok {
				cid = string(valBytes)
			}

			ctx := context.WithValue(context.Background(), CorrelationIDContextKey, cid)

			slog.InfoContext(ctx, "Event received from RabbitMQ", "queue", queueName, "action", event.Action, "id", event.ID)
			if err := handler(ctx, event); err != nil {
				slog.ErrorContext(ctx, "failed to handle message", "err", err)
				// Negative acknowledge without requeuing (routes to DLQ)
				_ = d.Nack(false, false)
			} else {
				// Manually acknowledge successfully processed message
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}
