package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	docspersist "github.com/Dovud1997/Dovud/backend/internal/modules/documents/infrastructure/persistence"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	notifydomain "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/domain"
	notifypersist "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/infrastructure/persistence"
	tenantapp "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	tenantpersist "github.com/Dovud1997/Dovud/backend/internal/modules/tenant/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/crypto"
	"github.com/Dovud1997/Dovud/backend/internal/platform/database"
	"github.com/Dovud1997/Dovud/backend/internal/platform/logger"
	"github.com/Dovud1997/Dovud/backend/internal/platform/media"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/rabbitmqx"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Runner struct {
	cfg      *config.Config
	log      *slog.Logger
	db       *gorm.DB
	outbox   *outbox.Store
	mq       *rabbitmqx.Client
	store    storage.ObjectStore
	notify    *notify.Router
	files     *docspersist.FileRepo
	notifs    *notifypersist.NotificationRepo
	users     *identitypersist.UserRepo
	devices   *identitypersist.DeviceRepo
	tenantSvc *tenantapp.TenantService
	seen      sync.Map
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
	if err := db.AutoMigrate(
		&outbox.EventModel{},
		&docspersist.FileModel{},
		&notifypersist.NotificationModel{},
		&notifypersist.NotificationDeliveryModel{},
		&tenantpersist.ProviderModel{},
	); err != nil {
		return nil, err
	}
	mq, err := rabbitmqx.Connect(cfg.RabbitMQ.URL, log)
	if err != nil {
		return nil, err
	}
	store, err := storage.Open(cfg.Minio, cfg.App.PublicBaseURL, cfg.Auth.AccessSecret, log)
	if err != nil {
		log.Warn("worker storage unavailable", "error", err)
	}
	box, err := crypto.NewSecretBox(cfg.Auth.AccessSecret)
	if err != nil {
		return nil, err
	}
	tenantSvc := tenantapp.NewTenantService(
		tenantpersist.NewTenantRepo(db),
		tenantpersist.NewBrandingRepo(db),
		tenantpersist.NewDomainRepo(db),
	).WithProviders(tenantpersist.NewProviderRepo(db), box)
	return &Runner{
		cfg: cfg, log: log, db: db, outbox: outbox.NewStore(db), mq: mq,
		store: store, notify: notify.NewRouter(cfg.Notify, log),
		files: docspersist.NewFileRepo(db), notifs: notifypersist.NewNotificationRepo(db),
		users: identitypersist.NewUserRepo(db), devices: identitypersist.NewDeviceRepo(db),
		tenantSvc: tenantSvc,
	}, nil
}

func (r *Runner) Run() error {
	defer r.mq.Close()

	consumers := []struct {
		queue string
		tag   string
		fn    rabbitmqx.HandlerFunc
	}{
		{rabbitmqx.QueueNotifyEmail, "sfa-worker-email", r.handleNotify},
		{rabbitmqx.QueueNotifyPush, "sfa-worker-push", r.handleNotify},
		{rabbitmqx.QueueNotifySMS, "sfa-worker-sms", r.handleNotify},
		{rabbitmqx.QueueMediaProcess, "sfa-worker-media", r.handleMedia},
		{rabbitmqx.QueueAuditWrite, "sfa-worker-audit", r.handleAudit},
	}
	for _, c := range consumers {
		if err := r.mq.Consume(c.queue, c.tag, c.fn); err != nil {
			return err
		}
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
		if err := r.mq.Publish(ctx, ev.EventType, env); err != nil {
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
	_ = d
	if !r.once(env.EventID) {
		return nil
	}
	channel := strings.TrimPrefix(env.EventType, "notification.")
	if channel == env.EventType {
		channel = "email"
	}
	notifIDStr, _ := env.Payload["notification_id"].(string)
	userIDStr, _ := env.Payload["user_id"].(string)
	title, _ := env.Payload["title"].(string)
	body, _ := env.Payload["body"].(string)
	to := userIDStr
	var pushTargets []PushTarget
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			if tid, err := uuid.Parse(env.TenantID); err == nil {
				switch channel {
				case "email":
					if u, err := r.users.FindByID(ctx, tid, uid); err == nil {
						to = u.Email
					}
				case "push":
					if devices, err := r.devices.ListByUser(ctx, tid, uid); err == nil {
						pushTargets = CollectPushTargets(devices)
					}
				}
			}
		}
	}
	msg := notify.Message{To: to, Subject: title, Body: body, Data: env.Payload}
	sender := r.notify
	if r.tenantSvc != nil && env.TenantID != "" {
		if tid, err := uuid.Parse(env.TenantID); err == nil {
			cfg := r.tenantSvc.ResolveNotifyConfig(ctx, tid, r.cfg.Notify)
			sender = notify.NewRouter(cfg, r.log)
		}
	}

	var err error
	status := notifydomain.DeliverySent
	var errMsg *string
	var notifID uuid.UUID
	if notifIDStr != "" {
		notifID, _ = uuid.Parse(notifIDStr)
	}

	if channel == "push" {
		if len(pushTargets) == 0 {
			// No real device tokens (stubs only / none): keep single log-friendly send to user id.
			err = sender.Send(ctx, channel, msg)
			if notifID != uuid.Nil {
				st := notifydomain.DeliverySent
				var em *string
				if err != nil {
					st = notifydomain.DeliveryFailed
					s := err.Error()
					em = &s
				}
				_ = r.notifs.UpdateDeliveryStatus(ctx, notifID, channel, st, em)
			}
		} else {
			results, sent, failed, ferr := FanoutPush(ctx, sender, pushTargets, msg)
			err = ferr
			r.log.Info("push fan-out", "sent", sent, "failed", failed, "devices", len(pushTargets), "event_id", env.EventID)
			tokens := make([]string, 0, len(pushTargets))
			for _, t := range pushTargets {
				tokens = append(tokens, t.Token)
			}
			to = strings.Join(tokens, ",")
			if notifID != uuid.Nil {
				for _, res := range results {
					st := notifydomain.DeliverySent
					var em *string
					if res.Err != nil {
						st = notifydomain.DeliveryFailed
						s := res.Err.Error()
						em = &s
					}
					devID := res.Target.DeviceID
					plat := res.Target.Platform
					suf := TokenSuffix(res.Target.Token)
					_ = r.notifs.UpsertDeviceDelivery(ctx, &notifydomain.NotificationDelivery{
						NotificationID: notifID,
						Channel:        channel,
						Status:         st,
						Error:          em,
						DeviceID:       &devID,
						Platform:       &plat,
						TokenSuffix:    &suf,
					})
				}
				agg := notifydomain.DeliverySent
				var aggErr *string
				if err != nil {
					agg = notifydomain.DeliveryFailed
					s := err.Error()
					aggErr = &s
				}
				_ = r.notifs.UpdateDeliveryStatus(ctx, notifID, channel, agg, aggErr)
			}
		}
	} else {
		err = sender.Send(ctx, channel, msg)
		if err != nil {
			status = notifydomain.DeliveryFailed
			s := err.Error()
			errMsg = &s
			r.log.Error("notify send failed", "channel", channel, "error", err)
		}
		if notifID != uuid.Nil {
			_ = r.notifs.UpdateDeliveryStatus(ctx, notifID, channel, status, errMsg)
		}
	}

	if err != nil {
		if channel == "push" {
			r.log.Error("notify send failed", "channel", channel, "error", err)
		}
		return err
	}
	r.log.Info("notify delivered", "channel", channel, "to", to, "event_id", env.EventID)
	return nil
}

func (r *Runner) handleMedia(ctx context.Context, env rabbitmqx.Envelope, d amqp.Delivery) error {
	_ = d
	if !r.once(env.EventID) {
		return nil
	}
	if r.store == nil {
		r.log.Warn("skip media: no storage")
		return nil
	}
	mime, _ := env.Payload["mime"].(string)
	objectKey, _ := env.Payload["object_key"].(string)
	fileIDStr, _ := env.Payload["file_id"].(string)
	if objectKey == "" || !media.IsImageMIME(mime) {
		r.log.Info("media skip (not image)", "mime", mime, "object_key", objectKey)
		return nil
	}
	rc, err := r.store.Get(ctx, objectKey)
	if err != nil {
		return err
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	thumb, err := media.GenerateJPEGThumbnail(bytes.NewReader(raw), mime)
	if err != nil {
		r.log.Warn("thumbnail generate failed", "error", err)
		return nil
	}
	thumbKey := media.ThumbnailObjectKey(objectKey)
	if err := r.store.Put(ctx, thumbKey, "image/jpeg", bytes.NewReader(thumb), int64(len(thumb))); err != nil {
		return err
	}
	if fileIDStr != "" && env.TenantID != "" {
		if fid, err := uuid.Parse(fileIDStr); err == nil {
			if tid, err := uuid.Parse(env.TenantID); err == nil {
				if f, err := r.files.FindByID(ctx, tid, fid); err == nil {
					f.ThumbnailKey = &thumbKey
					meta := `{"thumbnail":"ready","width_max":256}`
					f.MetaJSON = &meta
					_ = r.files.Update(ctx, f)
				}
			}
		}
	}
	r.log.Info("thumbnail ready", "object_key", objectKey, "thumb_key", thumbKey, "bytes", len(thumb))
	return nil
}

func (r *Runner) handleAudit(ctx context.Context, env rabbitmqx.Envelope, d amqp.Delivery) error {
	_ = ctx
	_ = d
	if !r.once(env.EventID) {
		return nil
	}
	r.log.Info("audit event consumed", "event_type", env.EventType, "tenant_id", env.TenantID)
	return nil
}
