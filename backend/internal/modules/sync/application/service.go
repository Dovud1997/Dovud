package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	devices   domain.DeviceRepository
	changelog domain.ChangeLogRepository
	conflicts domain.ConflictRepository
}

func NewService(devices domain.DeviceRepository, changelog domain.ChangeLogRepository, conflicts domain.ConflictRepository) *Service {
	return &Service{devices: devices, changelog: changelog, conflicts: conflicts}
}

type BootstrapInput struct {
	DeviceID   string  `json:"device_id"`
	Platform   *string `json:"platform"`
	AppVersion *string `json:"app_version"`
}

type BootstrapResult struct {
	SyncProtocol    int               `json:"sync_protocol"`
	Cursors         map[string]string `json:"cursors"`
	BrandingVersion *int              `json:"branding_version,omitempty"`
	DeviceID        string            `json:"device_id"`
}

type ChangeDTO struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Version    int64  `json:"version"`
	Deleted    bool   `json:"deleted"`
	UpdatedAt  string `json:"updated_at"`
	Data       any    `json:"data"`
}

type PullResult struct {
	Changes    []ChangeDTO `json:"changes"`
	NextCursor string      `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
}

type PushInput struct {
	DeviceID string          `json:"device_id"`
	Ops      []domain.SyncOp `json:"ops"`
}

type PushOpResult struct {
	OpID       string     `json:"op_id"`
	Status     string     `json:"status"`
	Version    *int64     `json:"version,omitempty"`
	ConflictID *uuid.UUID `json:"conflict_id,omitempty"`
	Message    string     `json:"message,omitempty"`
}

type PushResult struct {
	Results []PushOpResult `json:"results"`
}

type ConflictDTO struct {
	ID            uuid.UUID  `json:"id"`
	DeviceID      string     `json:"device_id"`
	EntityType    string     `json:"entity_type"`
	EntityID      string     `json:"entity_id"`
	ClientOpID    string     `json:"client_op_id"`
	BaseVersion   int64      `json:"base_version"`
	ServerVersion int64      `json:"server_version"`
	ClientPayload any        `json:"client_payload"`
	ServerPayload any        `json:"server_payload"`
	Status        string     `json:"status"`
	Resolution    *string    `json:"resolution,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

type ResolveConflictInput struct {
	Resolution string `json:"resolution"`
}

type StatusResult struct {
	SyncProtocol   int        `json:"sync_protocol"`
	DeviceID       string     `json:"device_id"`
	LastPullCursor string     `json:"last_pull_cursor"`
	LastPullAt     *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	OpenConflicts  int64      `json:"open_conflicts"`
}

func parsePayload(raw string) any {
	if raw == "" {
		return map[string]any{}
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	return v
}

func marshalPayload(payload map[string]any) string {
	if payload == nil {
		return "{}"
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func toChangeDTO(c domain.SyncChange) ChangeDTO {
	return ChangeDTO{
		EntityType: c.EntityType,
		EntityID:   c.EntityID,
		Version:    c.Version,
		Deleted:    c.Deleted,
		UpdatedAt:  c.UpdatedAt.UTC().Format(time.RFC3339Nano),
		Data:       parsePayload(c.PayloadJSON),
	}
}

func toConflictDTO(c domain.SyncConflict) ConflictDTO {
	return ConflictDTO{
		ID: c.ID, DeviceID: c.DeviceID, EntityType: c.EntityType, EntityID: c.EntityID,
		ClientOpID: c.ClientOpID, BaseVersion: c.BaseVersion, ServerVersion: c.ServerVersion,
		ClientPayload: parsePayload(c.ClientPayload), ServerPayload: parsePayload(c.ServerPayload),
		Status: c.Status, Resolution: c.Resolution, CreatedAt: c.CreatedAt, ResolvedAt: c.ResolvedAt,
	}
}

func (s *Service) Bootstrap(ctx context.Context, tenantID, userID uuid.UUID, in BootstrapInput) (*BootstrapResult, error) {
	deviceID := strings.TrimSpace(in.DeviceID)
	if deviceID == "" {
		return nil, apperrors.ErrValidation
	}
	dev := &domain.SyncDevice{
		TenantID: tenantID, UserID: userID, DeviceID: deviceID,
		Platform: in.Platform, AppVersion: in.AppVersion,
	}
	if err := s.devices.Upsert(ctx, dev); err != nil {
		return nil, err
	}
	found, err := s.devices.Find(ctx, tenantID, userID, deviceID)
	if err != nil {
		return nil, err
	}
	return &BootstrapResult{
		SyncProtocol: domain.SyncProtocol,
		Cursors:      map[string]string{"default": found.LastPullCursor},
		DeviceID:     deviceID,
	}, nil
}

func (s *Service) Pull(ctx context.Context, tenantID, userID uuid.UUID, deviceID, cursor string, types []string, limit int) (*PullResult, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.devices.Find(ctx, tenantID, userID, deviceID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			dev := &domain.SyncDevice{TenantID: tenantID, UserID: userID, DeviceID: deviceID}
			if err := s.devices.Upsert(ctx, dev); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	changes, nextCursor, hasMore, err := s.changelog.ListSince(ctx, tenantID, cursor, types, limit)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_ = s.devices.UpdateCursors(ctx, tenantID, userID, deviceID, nextCursor, &now, nil)

	out := make([]ChangeDTO, 0, len(changes))
	for _, c := range changes {
		out = append(out, toChangeDTO(c))
	}
	return &PullResult{Changes: out, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s *Service) Push(ctx context.Context, tenantID, userID uuid.UUID, in PushInput) (*PushResult, error) {
	deviceID := strings.TrimSpace(in.DeviceID)
	if deviceID == "" {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.devices.Find(ctx, tenantID, userID, deviceID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			dev := &domain.SyncDevice{TenantID: tenantID, UserID: userID, DeviceID: deviceID}
			if err := s.devices.Upsert(ctx, dev); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	results := make([]PushOpResult, 0, len(in.Ops))
	for _, op := range in.Ops {
		results = append(results, s.applyOp(ctx, tenantID, userID, deviceID, op))
	}

	now := time.Now().UTC()
	_ = s.devices.UpdateCursors(ctx, tenantID, userID, deviceID, "", nil, &now)

	return &PushResult{Results: results}, nil
}

func (s *Service) applyOp(ctx context.Context, tenantID, userID uuid.UUID, deviceID string, op domain.SyncOp) PushOpResult {
	opID := strings.TrimSpace(op.OpID)
	entityType := strings.TrimSpace(op.EntityType)
	entityID := strings.TrimSpace(op.EntityID)
	opKind := strings.TrimSpace(strings.ToLower(op.Op))

	if opID == "" || entityType == "" || entityID == "" || !domain.ValidOp(opKind) {
		return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: "invalid op"}
	}

	applied, err := s.changelog.HasAppliedOp(ctx, tenantID, opID)
	if err != nil {
		return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
	}
	if applied {
		latest, lerr := s.changelog.FindLatest(ctx, tenantID, entityType, entityID)
		res := PushOpResult{OpID: opID, Status: domain.PushAcked, Message: "idempotent"}
		if lerr == nil {
			v := latest.Version
			res.Version = &v
		}
		return res
	}

	clientPayload := marshalPayload(op.Payload)

	switch opKind {
	case domain.OpCreate:
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: 1, Deleted: false, PayloadJSON: clientPayload,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		v := int64(1)
		return PushOpResult{OpID: opID, Status: domain.PushAcked, Version: &v}

	case domain.OpUpdate:
		latest, err := s.changelog.FindLatest(ctx, tenantID, entityType, entityID)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		if latest != nil && latest.Version != op.BaseVersion {
			conflict := &domain.SyncConflict{
				TenantID: tenantID, DeviceID: deviceID, UserID: userID,
				EntityType: entityType, EntityID: entityID, ClientOpID: opID,
				BaseVersion: op.BaseVersion, ServerVersion: latest.Version,
				ClientPayload: clientPayload, ServerPayload: latest.PayloadJSON,
				Status: domain.ConflictStatusOpen,
			}
			if err := s.conflicts.Create(ctx, conflict); err != nil {
				return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
			}
			return PushOpResult{OpID: opID, Status: domain.PushConflict, ConflictID: &conflict.ID, Message: "version mismatch"}
		}
		newVersion := op.BaseVersion + 1
		if latest == nil {
			newVersion = 1
		} else {
			newVersion = latest.Version + 1
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: newVersion, Deleted: false, PayloadJSON: clientPayload,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		return PushOpResult{OpID: opID, Status: domain.PushAcked, Version: &newVersion}

	case domain.OpDelete:
		latest, err := s.changelog.FindLatest(ctx, tenantID, entityType, entityID)
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		if latest != nil && op.BaseVersion > 0 && latest.Version != op.BaseVersion {
			conflict := &domain.SyncConflict{
				TenantID: tenantID, DeviceID: deviceID, UserID: userID,
				EntityType: entityType, EntityID: entityID, ClientOpID: opID,
				BaseVersion: op.BaseVersion, ServerVersion: latest.Version,
				ClientPayload: clientPayload, ServerPayload: latest.PayloadJSON,
				Status: domain.ConflictStatusOpen,
			}
			if err := s.conflicts.Create(ctx, conflict); err != nil {
				return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
			}
			return PushOpResult{OpID: opID, Status: domain.PushConflict, ConflictID: &conflict.ID, Message: "version mismatch"}
		}
		newVersion := int64(1)
		payload := "{}"
		if latest != nil {
			newVersion = latest.Version + 1
			payload = latest.PayloadJSON
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: newVersion, Deleted: true, PayloadJSON: payload,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		return PushOpResult{OpID: opID, Status: domain.PushAcked, Version: &newVersion}
	}

	return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: "unsupported op"}
}

func (s *Service) ListConflicts(ctx context.Context, tenantID uuid.UUID, deviceID string) ([]ConflictDTO, error) {
	rows, err := s.conflicts.ListOpen(ctx, tenantID, strings.TrimSpace(deviceID))
	if err != nil {
		return nil, err
	}
	out := make([]ConflictDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, toConflictDTO(c))
	}
	return out, nil
}

func (s *Service) ResolveConflict(ctx context.Context, tenantID, id uuid.UUID, in ResolveConflictInput) (*ConflictDTO, error) {
	resolution := strings.TrimSpace(in.Resolution)
	if resolution == "" {
		return nil, apperrors.ErrValidation
	}
	if err := s.conflicts.Resolve(ctx, tenantID, id, resolution); err != nil {
		return nil, err
	}
	c, err := s.conflicts.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toConflictDTO(*c)
	return &dto, nil
}

func (s *Service) Status(ctx context.Context, tenantID, userID uuid.UUID, deviceID string) (*StatusResult, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, apperrors.ErrValidation
	}
	dev, err := s.devices.Find(ctx, tenantID, userID, deviceID)
	if err != nil {
		return nil, err
	}
	open, err := s.conflicts.CountOpen(ctx, tenantID, deviceID)
	if err != nil {
		return nil, err
	}
	return &StatusResult{
		SyncProtocol:   domain.SyncProtocol,
		DeviceID:       deviceID,
		LastPullCursor: dev.LastPullCursor,
		LastPullAt:     dev.LastPullAt,
		LastPushAt:     dev.LastPushAt,
		OpenConflicts:  open,
	}, nil
}

// RecordChange appends a server-side change for other modules to fan out via sync pull.
func (s *Service) RecordChange(ctx context.Context, tenantID uuid.UUID, entityType, entityID string, version int64, deleted bool, payload any) error {
	entityType = strings.TrimSpace(entityType)
	entityID = strings.TrimSpace(entityID)
	if entityType == "" || entityID == "" {
		return apperrors.ErrValidation
	}
	if version < 1 {
		latest, err := s.changelog.FindLatest(ctx, tenantID, entityType, entityID)
		if err == nil {
			version = latest.Version + 1
		} else if errors.Is(err, apperrors.ErrNotFound) {
			version = 1
		} else {
			return err
		}
	}
	raw := "{}"
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return apperrors.ErrValidation
		}
		raw = string(b)
	}
	return s.changelog.Append(ctx, &domain.SyncChange{
		TenantID: tenantID, EntityType: entityType, EntityID: entityID,
		Version: version, Deleted: deleted, PayloadJSON: raw,
	})
}
