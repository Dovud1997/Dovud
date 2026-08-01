package application

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/analytics/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	kpi domain.KpiRepository
	db  *gorm.DB
}

func NewService(kpi domain.KpiRepository, db *gorm.DB) *Service {
	return &Service{kpi: kpi, db: db}
}

type KpiDefinitionDTO struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Unit        string     `json:"unit"`
}

type KpiSnapshotDTO struct {
	ID          uuid.UUID  `json:"id"`
	KpiCode     string     `json:"kpi_code"`
	Period      string     `json:"period"`
	PeriodStart time.Time  `json:"period_start"`
	ScopeType   string     `json:"scope_type"`
	ScopeID     *uuid.UUID `json:"scope_id,omitempty"`
	Value       float64    `json:"value"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toDefDTO(d domain.KpiDefinition) KpiDefinitionDTO {
	return KpiDefinitionDTO{
		ID: d.ID, TenantID: d.TenantID, Code: d.Code, Name: d.Name,
		Description: d.Description, Unit: d.Unit,
	}
}

func toSnapDTO(s domain.KpiSnapshot) KpiSnapshotDTO {
	return KpiSnapshotDTO{
		ID: s.ID, KpiCode: s.KpiCode, Period: s.Period, PeriodStart: s.PeriodStart,
		ScopeType: s.ScopeType, ScopeID: s.ScopeID, Value: s.Value, CreatedAt: s.CreatedAt,
	}
}

var defaultKPIs = []domain.KpiDefinition{
	{Code: "sales_amount", Name: "Sales Amount", Description: "Total sales amount", Unit: "money"},
	{Code: "visit_count", Name: "Visit Count", Description: "Number of visits", Unit: "count"},
	{Code: "order_count", Name: "Order Count", Description: "Number of orders", Unit: "count"},
}

func (s *Service) ensureDefaultKPIs(ctx context.Context) error {
	defs, err := s.kpi.ListDefinitions(ctx, uuid.Nil)
	if err != nil {
		return err
	}
	if len(defs) > 0 {
		return nil
	}
	for i := range defaultKPIs {
		d := defaultKPIs[i]
		if err := s.kpi.CreateDefinition(ctx, &d); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListKPIDefinitions(ctx context.Context, tenantID uuid.UUID) ([]KpiDefinitionDTO, error) {
	if err := s.ensureDefaultKPIs(ctx); err != nil {
		return nil, err
	}
	rows, err := s.kpi.ListDefinitions(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]KpiDefinitionDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDefDTO(r))
	}
	return out, nil
}

func (s *Service) ListKPISnapshots(ctx context.Context, tenantID uuid.UUID, code, period string, page, perPage int) ([]KpiSnapshotDTO, int64, error) {
	if period != "" && !domain.ValidPeriod(period) {
		return nil, 0, apperrors.ErrValidation
	}
	rows, total, err := s.kpi.ListSnapshots(ctx, tenantID, domain.SnapshotFilters{
		Code: strings.TrimSpace(code), Period: strings.TrimSpace(period),
	}, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]KpiSnapshotDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toSnapDTO(r))
	}
	return out, total, nil
}

func startOfDayUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) DashboardSummary(ctx context.Context, tenantID uuid.UUID) (*domain.DashboardSummary, error) {
	today := startOfDayUTC(time.Now())
	tomorrow := today.Add(24 * time.Hour)
	summary := &domain.DashboardSummary{}

	if err := s.db.WithContext(ctx).Table("orders").
		Where("tenant_id = ? AND ordered_at >= ? AND ordered_at < ? AND deleted_at IS NULL", tenantID, today, tomorrow).
		Count(&summary.OrdersToday).Error; err != nil {
		if !isMissingTable(err) {
			return nil, err
		}
	}

	_ = s.db.WithContext(ctx).Table("orders").
		Select("COALESCE(SUM(grand_total), 0)").
		Where("tenant_id = ? AND ordered_at >= ? AND ordered_at < ? AND deleted_at IS NULL", tenantID, today, tomorrow).
		Scan(&summary.OrdersTotalAmountToday)

	if err := s.db.WithContext(ctx).Table("visits").
		Where("tenant_id = ? AND started_at >= ? AND started_at < ? AND deleted_at IS NULL", tenantID, today, tomorrow).
		Count(&summary.VisitsToday).Error; err != nil {
		if !isMissingTable(err) {
			return nil, err
		}
	}

	if err := s.db.WithContext(ctx).Table("sales_agents").
		Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, "active").
		Count(&summary.ActiveAgents).Error; err != nil {
		if !isMissingTable(err) {
			return nil, err
		}
	}

	pendingStatuses := []string{"submitted", "confirmed", "picking", "shipped"}
	if err := s.db.WithContext(ctx).Table("orders").
		Where("tenant_id = ? AND status IN ? AND deleted_at IS NULL", tenantID, pendingStatuses).
		Count(&summary.PendingOrders).Error; err != nil {
		if !isMissingTable(err) {
			return nil, err
		}
	}

	summary.OpenReceivables = s.openReceivables(ctx, tenantID)
	return summary, nil
}

func (s *Service) openReceivables(ctx context.Context, tenantID uuid.UUID) float64 {
	if !s.db.Migrator().HasTable("receivables") {
		return 0
	}
	var total float64
	_ = s.db.WithContext(ctx).Table("receivables").
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status IN ?", tenantID, []string{"open", "partial"}).
		Scan(&total)
	return total
}

func isMissingTable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "undefined table")
}

func parseDay(s string) time.Time {
	s = strings.TrimSpace(s)
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return startOfDayUTC(t)
	}
	return time.Time{}
}

func (s *Service) SalesAnalytics(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.SalesPoint, error) {
	if to.Before(from) {
		return nil, apperrors.ErrValidation
	}
	type row struct {
		Day    string
		Orders int64
		Amount float64
	}
	var rows []row
	err := s.db.WithContext(ctx).Table("orders").
		Select("DATE(ordered_at) as day, COUNT(*) as orders, COALESCE(SUM(grand_total), 0) as amount").
		Where("tenant_id = ? AND ordered_at >= ? AND ordered_at < ? AND deleted_at IS NULL", tenantID, from, to).
		Group("DATE(ordered_at)").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		if isMissingTable(err) {
			return []domain.SalesPoint{}, nil
		}
		return nil, err
	}
	out := make([]domain.SalesPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.SalesPoint{Date: parseDay(r.Day), Orders: r.Orders, Amount: r.Amount})
	}
	return out, nil
}

func (s *Service) VisitsAnalytics(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.VisitPoint, error) {
	if to.Before(from) {
		return nil, apperrors.ErrValidation
	}
	type row struct {
		Day    string
		Visits int64
	}
	var rows []row
	err := s.db.WithContext(ctx).Table("visits").
		Select("DATE(started_at) as day, COUNT(*) as visits").
		Where("tenant_id = ? AND started_at >= ? AND started_at < ? AND deleted_at IS NULL", tenantID, from, to).
		Group("DATE(started_at)").
		Order("day ASC").
		Scan(&rows).Error
	if err != nil {
		if isMissingTable(err) {
			return []domain.VisitPoint{}, nil
		}
		return nil, err
	}
	out := make([]domain.VisitPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.VisitPoint{Date: parseDay(r.Day), Visits: r.Visits})
	}
	return out, nil
}
