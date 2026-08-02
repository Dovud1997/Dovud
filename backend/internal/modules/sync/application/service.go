package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/syncport"
	"github.com/google/uuid"
)

type DeviceLocker interface {
	TryLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key, token string) error
}

type LiveNotifier interface {
	BroadcastSyncInvalidate(tenantID uuid.UUID, entityTypes ...string)
	Publish(tenantID uuid.UUID, eventType string, payload map[string]any)
}

type Service struct {
	devices    domain.DeviceRepository
	changelog  domain.ChangeLogRepository
	conflicts  domain.ConflictRepository
	locker     DeviceLocker
	live       LiveNotifier
	applicator syncport.EntityApplicator
}

func NewService(devices domain.DeviceRepository, changelog domain.ChangeLogRepository, conflicts domain.ConflictRepository) *Service {
	return &Service{devices: devices, changelog: changelog, conflicts: conflicts}
}

func (s *Service) WithLocker(locker DeviceLocker) *Service {
	s.locker = locker
	return s
}

func (s *Service) WithLive(live LiveNotifier) *Service {
	s.live = live
	return s
}

func (s *Service) WithApplicator(app syncport.EntityApplicator) *Service {
	s.applicator = app
	return s
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
	Resolution    string         `json:"resolution"`
	MergedPayload map[string]any `json:"merged_payload,omitempty"`
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

	unlock := func() {}
	if s.locker != nil {
		lockKey := fmt.Sprintf("sync:lock:%s:%s", tenantID.String(), deviceID)
		token := uuid.NewString()
		ok, err := s.locker.TryLock(ctx, lockKey, token, 30*time.Second)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, apperrors.ErrRateLimited
		}
		unlock = func() { _ = s.locker.Unlock(ctx, lockKey, token) }
	}
	defer unlock()

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
		payloadJSON := clientPayload
		version := int64(1)
		deleted := false
		if s.applicator != nil && s.applicator.Supports(entityType) {
			applied, aerr := s.applicator.Apply(syncport.WithoutFanout(ctx), syncport.ApplyRequest{
				TenantID: tenantID, UserID: userID, EntityType: entityType, EntityID: entityID,
				Op: domain.OpCreate, Payload: op.Payload,
			})
			if aerr != nil {
				return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: aerr.Error()}
			}
			payloadJSON = marshalPayload(asMap(applied.Payload))
			version = applied.Version
			if version < 1 {
				version = 1
			}
			deleted = applied.Deleted
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: version, Deleted: deleted, PayloadJSON: payloadJSON,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		if s.live != nil {
			s.live.BroadcastSyncInvalidate(tenantID, entityType)
		}
		return PushOpResult{OpID: opID, Status: domain.PushAcked, Version: &version}

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
		payloadJSON := clientPayload
		if s.applicator != nil && s.applicator.Supports(entityType) {
			applied, aerr := s.applicator.Apply(syncport.WithoutFanout(ctx), syncport.ApplyRequest{
				TenantID: tenantID, UserID: userID, EntityType: entityType, EntityID: entityID,
				Op: domain.OpUpdate, Payload: op.Payload,
			})
			if aerr != nil {
				return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: aerr.Error()}
			}
			payloadJSON = marshalPayload(asMap(applied.Payload))
			if applied.Version > 0 {
				newVersion = applied.Version
			}
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: newVersion, Deleted: false, PayloadJSON: payloadJSON,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		if s.live != nil {
			s.live.BroadcastSyncInvalidate(tenantID, entityType)
		}
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
		if s.applicator != nil && s.applicator.Supports(entityType) {
			applied, aerr := s.applicator.Apply(syncport.WithoutFanout(ctx), syncport.ApplyRequest{
				TenantID: tenantID, UserID: userID, EntityType: entityType, EntityID: entityID,
				Op: domain.OpDelete, Payload: op.Payload,
			})
			if aerr != nil {
				return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: aerr.Error()}
			}
			payload = marshalPayload(asMap(applied.Payload))
			if applied.Version > 0 {
				newVersion = applied.Version
			}
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: entityType, EntityID: entityID,
			Version: newVersion, Deleted: true, PayloadJSON: payload,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: err.Error()}
		}
		_ = s.changelog.MarkAppliedOp(ctx, tenantID, opID)
		if s.live != nil {
			s.live.BroadcastSyncInvalidate(tenantID, entityType)
		}
		return PushOpResult{OpID: opID, Status: domain.PushAcked, Version: &newVersion}
	}

	return PushOpResult{OpID: opID, Status: domain.PushRejected, Message: "unsupported op"}
}

func asMap(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{}
	}
	return out
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
	switch resolution {
	case domain.ResolutionServerWins, domain.ResolutionClientWins, domain.ResolutionMerge:
	default:
		return nil, apperrors.ErrValidation
	}
	c, err := s.conflicts.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.ConflictStatusOpen {
		dto := toConflictDTO(*c)
		return &dto, nil
	}

	if resolution == domain.ResolutionClientWins || resolution == domain.ResolutionMerge {
		payloadMap := map[string]any{}
		if resolution == domain.ResolutionMerge {
			if len(in.MergedPayload) == 0 {
				return nil, apperrors.ErrValidation
			}
			payloadMap = in.MergedPayload
		} else if err := json.Unmarshal([]byte(c.ClientPayload), &payloadMap); err != nil {
			payloadMap = map[string]any{}
		}
		newVersion := c.ServerVersion + 1
		if newVersion < 1 {
			newVersion = 1
		}
		payloadJSON := marshalPayload(payloadMap)
		if s.applicator != nil && s.applicator.Supports(c.EntityType) {
			applyCtx := syncport.WithoutFanout(ctx)
			applied, aerr := s.applicator.Apply(applyCtx, syncport.ApplyRequest{
				TenantID: tenantID, UserID: c.UserID, EntityType: c.EntityType, EntityID: c.EntityID,
				Op: domain.OpUpdate, Payload: payloadMap,
			})
			if aerr != nil {
				// Fall back to create if missing.
				applied, aerr = s.applicator.Apply(applyCtx, syncport.ApplyRequest{
					TenantID: tenantID, UserID: c.UserID, EntityType: c.EntityType, EntityID: c.EntityID,
					Op: domain.OpCreate, Payload: payloadMap,
				})
			}
			if aerr != nil {
				return nil, aerr
			}
			payloadJSON = marshalPayload(asMap(applied.Payload))
			if applied.Version > 0 {
				newVersion = applied.Version
			}
		}
		change := &domain.SyncChange{
			TenantID: tenantID, EntityType: c.EntityType, EntityID: c.EntityID,
			Version: newVersion, Deleted: false, PayloadJSON: payloadJSON,
		}
		if err := s.changelog.Append(ctx, change); err != nil {
			return nil, err
		}
		if c.ClientOpID != "" {
			_ = s.changelog.MarkAppliedOp(ctx, tenantID, c.ClientOpID)
		}
		if s.live != nil {
			s.live.BroadcastSyncInvalidate(tenantID, c.EntityType)
		}
	}

	if err := s.conflicts.Resolve(ctx, tenantID, id, resolution); err != nil {
		return nil, err
	}
	c, err = s.conflicts.FindByID(ctx, tenantID, id)
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
	if err := s.changelog.Append(ctx, &domain.SyncChange{
		TenantID: tenantID, EntityType: entityType, EntityID: entityID,
		Version: version, Deleted: deleted, PayloadJSON: raw,
	}); err != nil {
		return err
	}
	if s.live != nil {
		s.live.BroadcastSyncInvalidate(tenantID, entityType)
		if evt := domainLiveEvent(entityType); evt != "" {
			s.live.Publish(tenantID, evt, map[string]any{
				"entity_type": entityType,
				"entity_id":   entityID,
				"version":     version,
				"deleted":     deleted,
			})
		}
	}
	return nil
}

func domainLiveEvent(entityType string) string {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "order":
		return "order.updated"
	case "visit":
		return "visit.updated"
	case "product", "product_price":
		return "product.updated"
	case "notification":
		return "notification.created"
	default:
		return ""
	}
}
