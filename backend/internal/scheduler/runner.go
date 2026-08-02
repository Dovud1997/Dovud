package scheduler

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	analyticsapp "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/application"
	analyticspersist "github.com/Dovud1997/Dovud/backend/internal/modules/analytics/infrastructure/persistence"
	identitypersist "github.com/Dovud1997/Dovud/backend/internal/modules/identity/infrastructure/persistence"
	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/database"
	"github.com/Dovud1997/Dovud/backend/internal/platform/logger"
	"github.com/Dovud1997/Dovud/backend/internal/platform/rabbitmqx"
	"github.com/google/uuid"
)

type Runner struct {
	cfg       *config.Config
	log       *slog.Logger
	tokens    *identitypersist.RefreshTokenRepo
	analytics *analyticsapp.Service
	mq        *rabbitmqx.Client
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
	_ = db.AutoMigrate(
		&identitypersist.RefreshTokenModel{},
		&analyticspersist.KpiDefinitionModel{},
		&analyticspersist.KpiSnapshotModel{},
	)

	var mq *rabbitmqx.Client
	if cfg.RabbitMQ.URL != "" {
		mq, err = rabbitmqx.Connect(cfg.RabbitMQ.URL, log)
		if err != nil {
			log.Warn("scheduler rabbitmq unavailable", "error", err)
		}
	}

	return &Runner{
		cfg:       cfg,
		log:       log,
		tokens:    identitypersist.NewRefreshTokenRepo(db),
		analytics: analyticsapp.NewService(analyticspersist.NewKpiRepo(db), db),
		mq:        mq,
	}, nil
}

func (r *Runner) Run() error {
	if r.mq != nil {
		defer r.mq.Close()
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	r.log.Info("scheduler started", "env", r.cfg.App.Env)

	tokenTicker := time.NewTicker(15 * time.Minute)
	kpiTicker := time.NewTicker(30 * time.Minute)
	outboxTicker := time.NewTicker(10 * time.Second)
	defer tokenTicker.Stop()
	defer kpiTicker.Stop()
	defer outboxTicker.Stop()

	// Run once on boot.
	r.cleanupTokens(ctx)
	r.recomputeKPI(ctx)
	r.nudgeOutbox(ctx)

	for {
		select {
		case <-ctx.Done():
			r.log.Info("scheduler shutting down")
			return nil
		case <-tokenTicker.C:
			r.cleanupTokens(ctx)
		case <-kpiTicker.C:
			r.recomputeKPI(ctx)
		case <-outboxTicker.C:
			r.nudgeOutbox(ctx)
		}
	}
}

func (r *Runner) cleanupTokens(ctx context.Context) {
	n, err := r.tokens.DeleteExpired(ctx, time.Now().UTC())
	if err != nil {
		r.log.Error("token cleanup failed", "error", err)
		return
	}
	if n > 0 {
		r.log.Info("token cleanup", "deleted", n)
	}
}

func (r *Runner) recomputeKPI(ctx context.Context) {
	n, err := r.analytics.RecomputeDailySnapshots(ctx)
	if err != nil {
		r.log.Error("kpi recompute failed", "error", err)
		return
	}
	r.log.Info("kpi snapshots written", "count", n)
}

func (r *Runner) nudgeOutbox(ctx context.Context) {
	if r.mq == nil {
		return
	}
	// Wake workers; they also poll outbox on their own ticker.
	_ = r.mq.Publish(ctx, "outbox.publish", rabbitmqx.Envelope{
		EventID: uuid.NewString(), EventType: "outbox.publish",
		TenantID: uuid.Nil.String(), OccurredAt: time.Now().UTC(),
		Payload: map[string]any{"source": "scheduler"},
	})
}
