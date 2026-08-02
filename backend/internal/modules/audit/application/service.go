package application

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/google/uuid"
)

type Service struct {
	repo   domain.AuditRepository
	outbox *outbox.Store
}

func NewService(repo domain.AuditRepository, outboxStore *outbox.Store) *Service {
	return &Service{repo: repo, outbox: outboxStore}
}

type AuditLogDTO struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	ActorUserID *uuid.UUID `json:"actor_user_id,omitempty"`
	Action      string     `json:"action"`
	EntityType  *string    `json:"entity_type,omitempty"`
	EntityID    *uuid.UUID `json:"entity_id,omitempty"`
	BeforeJSON  any        `json:"before,omitempty"`
	AfterJSON   any        `json:"after,omitempty"`
	IP          *string    `json:"ip,omitempty"`
	UserAgent   *string    `json:"user_agent,omitempty"`
	RequestID   *string    `json:"request_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type WriteInput struct {
	TenantID    *uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	Before      any
	After       any
	IP          string
	UserAgent   string
	RequestID   string
}

func parseJSONPtr(raw *string) any {
	if raw == nil || *raw == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(*raw), &v); err != nil {
		return *raw
	}
	return v
}

func toDTO(l domain.AuditLog) AuditLogDTO {
	return AuditLogDTO{
		ID: l.ID, TenantID: l.TenantID, ActorUserID: l.ActorUserID, Action: l.Action,
		EntityType: l.EntityType, EntityID: l.EntityID,
		BeforeJSON: parseJSONPtr(l.BeforeJSON), AfterJSON: parseJSONPtr(l.AfterJSON),
		IP: l.IP, UserAgent: l.UserAgent, RequestID: l.RequestID, CreatedAt: l.CreatedAt,
	}
}

func marshalOpt(v any) *string {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func (s *Service) Write(ctx context.Context, in WriteInput) (*AuditLogDTO, error) {
	action := strings.TrimSpace(in.Action)
	if action == "" {
		return nil, apperrors.ErrValidation
	}
	var entityType *string
	if t := strings.TrimSpace(in.EntityType); t != "" {
		entityType = &t
	}
	var ip, ua, reqID *string
	if in.IP != "" {
		ip = &in.IP
	}
	if in.UserAgent != "" {
		ua = &in.UserAgent
	}
	if in.RequestID != "" {
		reqID = &in.RequestID
	}
	row := &domain.AuditLog{
		TenantID: in.TenantID, ActorUserID: in.ActorUserID, Action: action,
		EntityType: entityType, EntityID: in.EntityID,
		BeforeJSON: marshalOpt(in.Before), AfterJSON: marshalOpt(in.After),
		IP: ip, UserAgent: ua, RequestID: reqID,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	if s.outbox != nil && in.TenantID != nil {
		_ = s.outbox.Append(ctx, *in.TenantID, "audit", &row.ID, "audit.written", map[string]any{
			"audit_id": row.ID.String(), "action": row.Action,
		})
	}
	dto := toDTO(*row)
	return &dto, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*AuditLogDTO, error) {
	row, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toDTO(*row)
	return &dto, nil
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filters domain.ListFilters, page, perPage int) ([]AuditLogDTO, int64, error) {
	rows, total, err := s.repo.List(ctx, tenantID, filters, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]AuditLogDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDTO(r))
	}
	return out, total, nil
}
