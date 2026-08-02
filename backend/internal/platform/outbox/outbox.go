package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusPending   = "pending"
	StatusPublished = "published"
	StatusFailed    = "failed"
)

type Event struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	AggregateType string
	AggregateID   *uuid.UUID
	EventType     string
	PayloadJSON   string
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

type EventModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	AggregateType string     `gorm:"size:64;not null"`
	AggregateID   *uuid.UUID `gorm:"type:uuid;index"`
	EventType     string     `gorm:"size:128;not null;index"`
	PayloadJSON   string     `gorm:"type:text;not null"`
	Status        string     `gorm:"size:32;not null;index"`
	Attempts      int        `gorm:"not null;default:0"`
	NextAttemptAt time.Time  `gorm:"not null;index"`
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

func (EventModel) TableName() string { return "outbox_events" }

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) Append(ctx context.Context, tenantID uuid.UUID, aggregateType string, aggregateID *uuid.UUID, eventType string, payload any) error {
	raw := "{}"
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		raw = string(b)
	}
	now := time.Now().UTC()
	m := EventModel{
		ID: uuid.New(), TenantID: tenantID, AggregateType: aggregateType, AggregateID: aggregateID,
		EventType: eventType, PayloadJSON: raw, Status: StatusPending, Attempts: 0,
		NextAttemptAt: now, CreatedAt: now,
	}
	return s.db.WithContext(ctx).Create(&m).Error
}

func (s *Store) AppendTx(ctx context.Context, tx *gorm.DB, tenantID uuid.UUID, aggregateType string, aggregateID *uuid.UUID, eventType string, payload any) error {
	db := s.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}
	raw := "{}"
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		raw = string(b)
	}
	now := time.Now().UTC()
	m := EventModel{
		ID: uuid.New(), TenantID: tenantID, AggregateType: aggregateType, AggregateID: aggregateID,
		EventType: eventType, PayloadJSON: raw, Status: StatusPending, Attempts: 0,
		NextAttemptAt: now, CreatedAt: now,
	}
	return db.Create(&m).Error
}

func toEvent(m EventModel) Event {
	return Event{
		ID: m.ID, TenantID: m.TenantID, AggregateType: m.AggregateType, AggregateID: m.AggregateID,
		EventType: m.EventType, PayloadJSON: m.PayloadJSON, Status: m.Status, Attempts: m.Attempts,
		NextAttemptAt: m.NextAttemptAt, CreatedAt: m.CreatedAt, PublishedAt: m.PublishedAt,
	}
}

func (s *Store) ClaimPending(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}
	now := time.Now().UTC()
	var rows []EventModel
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("status = ? AND next_attempt_at <= ?", StatusPending, now).
			Order("created_at ASC").
			Limit(limit).
			Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			rows[i].Attempts++
			rows[i].NextAttemptAt = now.Add(30 * time.Second)
			if err := tx.Model(&EventModel{}).Where("id = ?", rows[i].ID).Updates(map[string]any{
				"attempts": rows[i].Attempts, "next_attempt_at": rows[i].NextAttemptAt,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]Event, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEvent(m))
	}
	return out, nil
}

func (s *Store) MarkPublished(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Model(&EventModel{}).Where("id = ?", id).Updates(map[string]any{
		"status": StatusPublished, "published_at": now,
	}).Error
}

func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, retryAfter time.Duration) error {
	if retryAfter <= 0 {
		retryAfter = time.Minute
	}
	return s.db.WithContext(ctx).Model(&EventModel{}).Where("id = ?", id).Updates(map[string]any{
		"status": StatusPending, "next_attempt_at": time.Now().UTC().Add(retryAfter),
	}).Error
}
