package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/database"
	"github.com/Dovud1997/Dovud/backend/internal/platform/logger"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/rabbitmqx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Runner struct {
	cfg    *config.Config
	log    *slog.Logger
	outbox *outbox.Store
	mq     *rabbitmqx.Client
	seen   sync.Map // event_id -> struct{} for consumer idempotency
}

func New(cfgPath string) (*Runner, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.App.Env)
	db, err := database.Connect(cfg.Database.DSN, cfg.App.Env)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&outbox.EventModel{}); err != nil {
		return nil, err
	}
	mq, err := rabbitmqx.Connect(cfg.RabbitMQ.URL, log)
	if err != nil {
		return nil, err
	}
	return &Runner{
		cfg: cfg, log: log, outbox: outbox.NewStore(db), mq: mq,
	}, nil
}

func (r *Runner) Run() error {
	defer r.mq.Close()

	if err := r.mq.Consume(rabbitmqx.QueueNotifyEmail, "sfa-worker-email", r.handleNotify); err != nil {
		return err
	}
	if err := r.mq.Consume(rabbitmqx.QueueNotifyPush, "sfa-worker-push", r.handleNotify); err != nil {
		return err
	}
	if err := r.mq.Consume(rabbitmqx.QueueMediaProcess, "sfa-worker-media", r.handleMedia); err != nil {
		return err
	}
	if err := r.mq.Consume(rabbitmqx.QueueAuditWrite, "sfa-worker-audit", r.handleAudit); err != nil {
		return err
	}
	if err := r.mq.Consume(rabbitmqx.QueueNotifySMS, "sfa-worker-sms", r.handleNotify); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	r.log.Info("worker started", "env", r.cfg.App.Env)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info("worker shutting down")
			return nil
		case <-ticker.C:
			if err := r.relayOutbox(ctx); err != nil {
				r.log.Error("outbox relay failed", "error", err)
			}
		}
	}
}

func (r *Runner) relayOutbox(ctx context.Context) error {
	events, err := r.outbox.ClaimPending(ctx, 50)
	if err != nil {
		return err
	}
	for _, ev := range events {
		payload := map[string]any{}
		_ = json.Unmarshal([]byte(ev.PayloadJSON), &payload)
		aggID := ""
		if ev.AggregateID != nil {
			aggID = ev.AggregateID.String()
		}
		env := rabbitmqx.Envelope{
			EventID: ev.ID.String(), EventType: ev.EventType,
			TenantID: ev.TenantID.String(), AggregateType: ev.AggregateType,
			AggregateID: aggID, OccurredAt: ev.CreatedAt, Payload: payload,
		}
		routingKey := ev.EventType
		if stringsHasPrefix(ev.EventType, "media.") {
			routingKey = ev.EventType
		}
		if err := r.mq.Publish(ctx, routingKey, env); err != nil {
			_ = r.outbox.MarkFailed(ctx, ev.ID, time.Minute)
			return err
		}
		if err := r.outbox.MarkPublished(ctx, ev.ID); err != nil {
			return err
		}
		r.log.Info("outbox published", "event_id", ev.ID, "event_type", ev.EventType)
	}
	return nil
}

func (r *Runner) once(eventID string) bool {
	if eventID == "" {
		return true
	}
	if _, loaded := r.seen.LoadOrStore(eventID, struct{}{}); loaded {
		return false
	}
	return true
}

func (r *Runner) handleNotify(ctx context.Context, env rabbitmqx.Envelope, d amqp.Delivery) error {
	_ = ctx
	_ = d
	if !r.once(env.EventID) {
		r.log.Info("skip duplicate notify", "event_id", env.EventID)
		return nil
	}
	r.log.Info("notify handled", "event_type", env.EventType, "tenant_id", env.TenantID, "payload", env.Payload)
	return nil
}

func (r *Runner) handleMedia(ctx context.Context, env rabbitmqx.Envelope, d amqp.Delivery) error {
	_ = ctx
	_ = d
	if !r.once(env.EventID) {
		r.log.Info("skip duplicate media", "event_id", env.EventID)
		return nil
	}
	// Placeholder for thumbnail generation / virus scan hooks.
	mime, _ := env.Payload["mime"].(string)
	objectKey, _ := env.Payload["object_key"].(string)
	r.log.Info("media process handled",
		"event_type", env.EventType,
		"tenant_id", env.TenantID,
		"mime", mime,
		"object_key", objectKey,
		"thumbnails", "queued",
	)
	return nil
}

func (r *Runner) handleAudit(ctx context.Context, env rabbitmqx.Envelope, d amqp.Delivery) error {
	_ = ctx
	_ = d
	if !r.once(env.EventID) {
		return nil
	}
	r.log.Info("audit event consumed", "event_type", env.EventType, "tenant_id", env.TenantID, "payload", env.Payload)
	return nil
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
