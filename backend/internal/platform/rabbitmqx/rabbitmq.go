package rabbitmqx

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeEvents = "sfa.events"
	ExchangeDLX    = "sfa.dlx"

	QueueNotifyEmail  = "q.notify.email"
	QueueNotifyPush   = "q.notify.push"
	QueueNotifySMS    = "q.notify.sms"
	QueueMediaProcess = "q.media.process"
	QueueAuditWrite   = "q.audit.write"
	QueueOutboxRelay  = "q.outbox.relay"
	QueueDLQ          = "q.dlq"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *slog.Logger
}

func Connect(url string, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}
	if url == "" {
		return nil, fmt.Errorf("rabbitmq url empty")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	c := &Client{conn: conn, ch: ch, log: log}
	if err := c.DeclareTopology(); err != nil {
		_ = c.Close()
		return nil, err
	}
	log.Info("rabbitmq connected")
	return c, nil
}

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Channel() *amqp.Channel { return c.ch }

func (c *Client) DeclareTopology() error {
	if err := c.ch.ExchangeDeclare(ExchangeEvents, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.ch.ExchangeDeclare(ExchangeDLX, "fanout", true, false, false, false, nil); err != nil {
		return err
	}

	queues := []struct {
		name       string
		bindingKey string
	}{
		{QueueNotifyEmail, "notification.email"},
		{QueueNotifyPush, "notification.push"},
		{QueueNotifySMS, "notification.sms"},
		{QueueMediaProcess, "media.#"},
		{QueueAuditWrite, "audit.#"},
		{QueueOutboxRelay, "outbox.publish"},
	}

	args := amqp.Table{"x-dead-letter-exchange": ExchangeDLX}
	for _, q := range queues {
		if _, err := c.ch.QueueDeclare(q.name, true, false, false, false, args); err != nil {
			return err
		}
		if err := c.ch.QueueBind(q.name, q.bindingKey, ExchangeEvents, false, nil); err != nil {
			return err
		}
	}
	if _, err := c.ch.QueueDeclare(QueueDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.ch.QueueBind(QueueDLQ, "", ExchangeDLX, false, nil); err != nil {
		return err
	}
	return nil
}

type Envelope struct {
	EventID       string         `json:"event_id"`
	EventType     string         `json:"event_type"`
	TenantID      string         `json:"tenant_id"`
	AggregateType string         `json:"aggregate_type,omitempty"`
	AggregateID   string         `json:"aggregate_id,omitempty"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Payload       map[string]any `json:"payload,omitempty"`
}

func (c *Client) Publish(ctx context.Context, routingKey string, env Envelope) error {
	if env.OccurredAt.IsZero() {
		env.OccurredAt = time.Now().UTC()
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return c.ch.PublishWithContext(ctx, ExchangeEvents, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    env.EventID,
		Type:         env.EventType,
		Timestamp:    env.OccurredAt,
		Body:         body,
	})
}

type HandlerFunc func(ctx context.Context, env Envelope, delivery amqp.Delivery) error

func (c *Client) Consume(queue string, consumerTag string, handler HandlerFunc) error {
	deliveries, err := c.ch.Consume(queue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			var env Envelope
			if err := json.Unmarshal(d.Body, &env); err != nil {
				c.log.Error("invalid message", "queue", queue, "error", err)
				_ = d.Nack(false, false)
				continue
			}
			if env.EventType == "" {
				env.EventType = d.Type
			}
			if env.EventID == "" {
				env.EventID = d.MessageId
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := handler(ctx, env, d)
			cancel()
			if err != nil {
				c.log.Error("handler failed", "queue", queue, "event_id", env.EventID, "error", err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()
	return nil
}
