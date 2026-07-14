package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransactionEvent struct {
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

// Publish sends a message to the specified queue
func Publish(conn *amqp.Connection, queueName string, event TransactionEvent) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Info("Event published to RabbitMQ", "queue", queueName, "action", event.Action)
	return nil
}

// Consume runs a background loop to process messages from a queue
func Consume(conn *amqp.Connection, queueName string, handler func(TransactionEvent) error) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		defer ch.Close()
		slog.Info("Started consuming events from RabbitMQ", "queue", queueName)
		for d := range msgs {
			var event TransactionEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				slog.Error("failed to unmarshal message payload", "err", err)
				continue
			}

			slog.Info("Event received from RabbitMQ", "queue", queueName, "action", event.Action)
			if err := handler(event); err != nil {
				slog.Error("failed to handle message", "err", err)
			}
		}
	}()

	return nil
}
