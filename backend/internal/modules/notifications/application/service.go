package application

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/notifications/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/google/uuid"
)

type Service struct {
	repo   domain.NotificationRepository
	outbox *outbox.Store
}

func NewService(repo domain.NotificationRepository, outboxStore *outbox.Store) *Service {
	return &Service{repo: repo, outbox: outboxStore}
}

type NotificationDTO struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	PayloadJSON *string    `json:"payload_json,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	Channel     string     `json:"channel"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DeliveryDTO struct {
	ID             uuid.UUID `json:"id"`
	NotificationID uuid.UUID `json:"notification_id"`
	Channel        string    `json:"channel"`
	Status         string    `json:"status"`
	Error          *string   `json:"error,omitempty"`
	DeviceID       *string   `json:"device_id,omitempty"`
	Platform       *string   `json:"platform,omitempty"`
	TokenSuffix    *string   `json:"token_suffix,omitempty"`
	AttemptedAt    time.Time `json:"attempted_at"`
}

type CreateInput struct {
	UserID      uuid.UUID `json:"user_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PayloadJSON *string   `json:"payload_json"`
	Channel     string    `json:"channel"`
}

func toDTO(n domain.Notification) NotificationDTO {
	return NotificationDTO{
		ID: n.ID, UserID: n.UserID, Type: n.Type, Title: n.Title, Body: n.Body,
		PayloadJSON: n.PayloadJSON, ReadAt: n.ReadAt, Channel: n.Channel, CreatedAt: n.CreatedAt,
	}
}

func toDeliveryDTO(d domain.NotificationDelivery) DeliveryDTO {
	return DeliveryDTO{
		ID: d.ID, NotificationID: d.NotificationID, Channel: d.Channel,
		Status: d.Status, Error: d.Error, DeviceID: d.DeviceID, Platform: d.Platform,
		TokenSuffix: d.TokenSuffix, AttemptedAt: d.AttemptedAt,
	}
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, in CreateInput) (*NotificationDTO, error) {
	if in.UserID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	typ := strings.TrimSpace(in.Type)
	if title == "" || body == "" || typ == "" {
		return nil, apperrors.ErrValidation
	}
	channel := strings.TrimSpace(in.Channel)
	if channel == "" {
		channel = domain.ChannelInApp
	}
	if !domain.ValidChannel(channel) {
		return nil, apperrors.ErrValidation
	}

	n := &domain.Notification{
		TenantID: tenantID, UserID: in.UserID, Type: typ,
		Title: title, Body: body, PayloadJSON: in.PayloadJSON, Channel: channel,
	}
	deliveryStatus := domain.DeliverySent
	if channel != domain.ChannelInApp {
		deliveryStatus = domain.DeliveryPending
	}
	delivery := &domain.NotificationDelivery{
		Channel: channel, Status: deliveryStatus,
	}
	if err := s.repo.Create(ctx, n, delivery); err != nil {
		return nil, err
	}
	if s.outbox != nil && channel != domain.ChannelInApp {
		eventType := "notification." + channel
		_ = s.outbox.Append(ctx, tenantID, "notification", &n.ID, eventType, map[string]any{
			"notification_id": n.ID.String(),
			"user_id":         n.UserID.String(),
			"type":            n.Type,
			"title":           n.Title,
			"body":            n.Body,
			"channel":         channel,
		})
	}
	dto := toDTO(*n)
	return &dto, nil
}

func (s *Service) CreateTest(ctx context.Context, tenantID, userID uuid.UUID) (*NotificationDTO, error) {
	return s.Create(ctx, tenantID, CreateInput{
		UserID:  userID,
		Type:    "test",
		Title:   "Test notification",
		Body:    "This is a test in-app notification.",
		Channel: domain.ChannelInApp,
	})
}

func (s *Service) ListByUser(ctx context.Context, tenantID, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]NotificationDTO, int64, error) {
	rows, total, err := s.repo.ListByUser(ctx, tenantID, userID, domain.ListFilters{UnreadOnly: unreadOnly}, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]NotificationDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDTO(r))
	}
	return out, total, nil
}

func (s *Service) MarkRead(ctx context.Context, tenantID, userID, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, tenantID, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return s.repo.MarkAllRead(ctx, tenantID, userID)
}

func (s *Service) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return s.repo.UnreadCount(ctx, tenantID, userID)
}

func (s *Service) ListDeliveries(ctx context.Context, tenantID, notificationID uuid.UUID) ([]DeliveryDTO, error) {
	if _, err := s.repo.FindByID(ctx, tenantID, notificationID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListDeliveries(ctx, notificationID)
	if err != nil {
		return nil, err
	}
	out := make([]DeliveryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDeliveryDTO(r))
	}
	return out, nil
}
