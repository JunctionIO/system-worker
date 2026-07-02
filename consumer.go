package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	statusQueue = "junction.status"
	maxRetries  = 3
)

var baseBackoff = 5 * time.Second

type Consumer struct {
	conn   *amqp.Connection
	client *APIClient
}

func NewConsumer(conn *amqp.Connection, client *APIClient) *Consumer {
	return &Consumer{conn: conn, client: client}
}

func (c *Consumer) Run() error {
	ch, err := c.conn.Channel()

	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	defer ch.Close()

	_, err = ch.QueueDeclare(statusQueue, true, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	msgs, err := ch.Consume(statusQueue, "", false, false, false, false, nil)

	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	slog.Info("consuming", "queue", statusQueue)

	for msg := range msgs {
		c.process(msg)
	}

	return nil
}

func (c *Consumer) process(msg amqp.Delivery) {
	var status StatusMessage

	if err := json.Unmarshal(msg.Body, &status); err != nil {
		slog.Error("malformed message, discarding", "error", err)
		msg.Ack(false)
		return
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := c.client.PostStatus(status); err != nil {
			if attempt < maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * baseBackoff
				slog.Warn("status update failed, retrying",
					"trace_id", status.TraceID,
					"log_id", status.LogID,
					"attempt", attempt,
					"backoff", backoff,
					"error", err,
				)
				time.Sleep(backoff)
				continue
			}

			slog.Error("status update failed after all retries, discarding",
				"trace_id", status.TraceID,
				"log_id", status.LogID,
				"error", err,
			)

			msg.Ack(false)

			return
		}

		msg.Ack(false)

		return
	}
}
