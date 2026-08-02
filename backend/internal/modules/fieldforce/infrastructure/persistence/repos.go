package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- SalesAgent ---

type AgentRepo struct{ db *gorm.DB }

func NewAgentRepo(db *gorm.DB) *AgentRepo { return &AgentRepo{db: db} }

func toAgent(m SalesAgentModel) domain.SalesAgent {
	return domain.SalesAgent{
		ID: m.ID, TenantID: m.TenantID, UserID: m.UserID, BranchID: m.BranchID,
		EmployeeCode: m.EmployeeCode, ManagerID: m.ManagerID, Status: m.Status,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *AgentRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.SalesAgent, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&SalesAgentModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []SalesAgentModel
	if err := q.Order("employee_code").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.SalesAgent, 0, len(rows))
	for _, m := range rows {
		out = append(out, toAgent(m))
	}
	return out, total, nil
}

func (r *AgentRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.SalesAgent, error) {
	var m SalesAgentModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	a := toAgent(m)
	return &a, nil
}

func (r *AgentRepo) FindByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.SalesAgent, error) {
	var m SalesAgentModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	a := toAgent(m)
	return &a, nil
}

func (r *AgentRepo) Create(ctx context.Context, a *domain.SalesAgent) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now().UTC()
	a.CreatedAt, a.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&SalesAgentModel{
		ID: a.ID, TenantID: a.TenantID, UserID: a.UserID, BranchID: a.BranchID,
		EmployeeCode: a.EmployeeCode, ManagerID: a.ManagerID, Status: a.Status,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}).Error
}

func (r *AgentRepo) Update(ctx context.Context, a *domain.SalesAgent) error {
	a.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&SalesAgentModel{}).Where("id = ? AND tenant_id = ?", a.ID, a.TenantID).Updates(map[string]any{
		"user_id": a.UserID, "branch_id": a.BranchID, "employee_code": a.EmployeeCode,
		"manager_id": a.ManagerID, "status": a.Status, "updated_at": a.UpdatedAt,
	}).Error
}

func (r *AgentRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&SalesAgentModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// --- Route ---

type RouteRepo struct{ db *gorm.DB }

func NewRouteRepo(db *gorm.DB) *RouteRepo { return &RouteRepo{db: db} }

func toRoute(m RouteModel) domain.Route {
	return domain.Route{
		ID: m.ID, TenantID: m.TenantID, AgentID: m.AgentID, Date: m.Date,
		Name: m.Name, Status: m.Status, Version: m.Version,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func toStop(m RouteStopModel) domain.RouteStop {
	return domain.RouteStop{
		ID: m.ID, RouteID: m.RouteID, CustomerID: m.CustomerID, Sequence: m.Sequence,
		PlannedArrival: m.PlannedArrival, Status: m.Status,
	}
}

func (r *RouteRepo) List(ctx context.Context, tenantID uuid.UUID, agentID *uuid.UUID, date *time.Time, page, perPage int) ([]domain.Route, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&RouteModel{}).Where("tenant_id = ?", tenantID)
	if agentID != nil {
		q = q.Where("agent_id = ?", *agentID)
	}
	if date != nil {
		q = q.Where("date = ?", date.Format("2006-01-02"))
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []RouteModel
	if err := q.Order("date DESC, name").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Route, 0, len(rows))
	for _, m := range rows {
		out = append(out, toRoute(m))
	}
	return out, total, nil
}

func (r *RouteRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Route, error) {
	var m RouteModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	rt := toRoute(m)
	return &rt, nil
}

func (r *RouteRepo) Create(ctx context.Context, rt *domain.Route) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	now := time.Now().UTC()
	rt.CreatedAt, rt.UpdatedAt = now, now
	if rt.Version == 0 {
		rt.Version = 1
	}
	return r.db.WithContext(ctx).Create(&RouteModel{
		ID: rt.ID, TenantID: rt.TenantID, AgentID: rt.AgentID, Date: rt.Date,
		Name: rt.Name, Status: rt.Status, Version: rt.Version,
		CreatedAt: rt.CreatedAt, UpdatedAt: rt.UpdatedAt,
	}).Error
}

func (r *RouteRepo) Update(ctx context.Context, rt *domain.Route) error {
	rt.UpdatedAt = time.Now().UTC()
	rt.Version++
	return r.db.WithContext(ctx).Model(&RouteModel{}).Where("id = ? AND tenant_id = ?", rt.ID, rt.TenantID).Updates(map[string]any{
		"agent_id": rt.AgentID, "date": rt.Date, "name": rt.Name, "status": rt.Status,
		"version": rt.Version, "updated_at": rt.UpdatedAt,
	}).Error
}

func (r *RouteRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&RouteModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *RouteRepo) ReplaceStops(ctx context.Context, routeID uuid.UUID, stops []domain.RouteStop) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("route_id = ?", routeID).Delete(&RouteStopModel{}).Error; err != nil {
			return err
		}
		if len(stops) == 0 {
			return nil
		}
		rows := make([]RouteStopModel, 0, len(stops))
		for i := range stops {
			s := &stops[i]
			if s.ID == uuid.Nil {
				s.ID = uuid.New()
			}
			s.RouteID = routeID
			if s.Status == "" {
				s.Status = "pending"
			}
			rows = append(rows, RouteStopModel{
				ID: s.ID, RouteID: s.RouteID, CustomerID: s.CustomerID, Sequence: s.Sequence,
				PlannedArrival: s.PlannedArrival, Status: s.Status,
			})
		}
		return tx.Create(&rows).Error
	})
}

func (r *RouteRepo) ListStops(ctx context.Context, routeID uuid.UUID) ([]domain.RouteStop, error) {
	var rows []RouteStopModel
	if err := r.db.WithContext(ctx).Where("route_id = ?", routeID).Order("sequence").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.RouteStop, 0, len(rows))
	for _, m := range rows {
		out = append(out, toStop(m))
	}
	return out, nil
}

// --- Visit ---

type VisitRepo struct{ db *gorm.DB }

func NewVisitRepo(db *gorm.DB) *VisitRepo { return &VisitRepo{db: db} }

func toVisit(m VisitModel) domain.Visit {
	return domain.Visit{
		ID: m.ID, TenantID: m.TenantID, AgentID: m.AgentID, CustomerID: m.CustomerID,
		RouteStopID: m.RouteStopID, StartedAt: m.StartedAt, EndedAt: m.EndedAt,
		CheckinLat: m.CheckinLat, CheckinLng: m.CheckinLng, CheckoutLat: m.CheckoutLat, CheckoutLng: m.CheckoutLng,
		Result: m.Result, Notes: m.Notes, Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *VisitRepo) List(ctx context.Context, tenantID uuid.UUID, agentID, customerID *uuid.UUID, page, perPage int) ([]domain.Visit, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&VisitModel{}).Where("tenant_id = ?", tenantID)
	if agentID != nil {
		q = q.Where("agent_id = ?", *agentID)
	}
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []VisitModel
	if err := q.Order("started_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Visit, 0, len(rows))
	for _, m := range rows {
		out = append(out, toVisit(m))
	}
	return out, total, nil
}

func (r *VisitRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Visit, error) {
	var m VisitModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	v := toVisit(m)
	return &v, nil
}

func (r *VisitRepo) Create(ctx context.Context, v *domain.Visit) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	now := time.Now().UTC()
	v.CreatedAt, v.UpdatedAt = now, now
	if v.Version == 0 {
		v.Version = 1
	}
	return r.db.WithContext(ctx).Create(&VisitModel{
		ID: v.ID, TenantID: v.TenantID, AgentID: v.AgentID, CustomerID: v.CustomerID,
		RouteStopID: v.RouteStopID, StartedAt: v.StartedAt, EndedAt: v.EndedAt,
		CheckinLat: v.CheckinLat, CheckinLng: v.CheckinLng, CheckoutLat: v.CheckoutLat, CheckoutLng: v.CheckoutLng,
		Result: v.Result, Notes: v.Notes, Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}).Error
}

func (r *VisitRepo) Update(ctx context.Context, v *domain.Visit) error {
	v.UpdatedAt = time.Now().UTC()
	v.Version++
	return r.db.WithContext(ctx).Model(&VisitModel{}).Where("id = ? AND tenant_id = ?", v.ID, v.TenantID).Updates(map[string]any{
		"agent_id": v.AgentID, "customer_id": v.CustomerID, "route_stop_id": v.RouteStopID,
		"started_at": v.StartedAt, "ended_at": v.EndedAt,
		"checkin_lat": v.CheckinLat, "checkin_lng": v.CheckinLng,
		"checkout_lat": v.CheckoutLat, "checkout_lng": v.CheckoutLng,
		"result": v.Result, "notes": v.Notes, "version": v.Version, "updated_at": v.UpdatedAt,
	}).Error
}

func (r *VisitRepo) AddPhoto(ctx context.Context, p *domain.VisitPhoto) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.CreatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Create(&VisitPhotoModel{
		ID: p.ID, VisitID: p.VisitID, FileURL: p.FileURL, Caption: p.Caption, CreatedAt: p.CreatedAt,
	}).Error
}

func (r *VisitRepo) ListPhotos(ctx context.Context, visitID uuid.UUID) ([]domain.VisitPhoto, error) {
	var rows []VisitPhotoModel
	if err := r.db.WithContext(ctx).Where("visit_id = ?", visitID).Order("created_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.VisitPhoto, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.VisitPhoto{
			ID: m.ID, VisitID: m.VisitID, FileURL: m.FileURL, Caption: m.Caption, CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}

func (r *VisitRepo) AddComment(ctx context.Context, c *domain.VisitComment) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Create(&VisitCommentModel{
		ID: c.ID, VisitID: c.VisitID, AuthorUserID: c.AuthorUserID, Body: c.Body, CreatedAt: c.CreatedAt,
	}).Error
}

func (r *VisitRepo) ListComments(ctx context.Context, visitID uuid.UUID) ([]domain.VisitComment, error) {
	var rows []VisitCommentModel
	if err := r.db.WithContext(ctx).Where("visit_id = ?", visitID).Order("created_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.VisitComment, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.VisitComment{
			ID: m.ID, VisitID: m.VisitID, AuthorUserID: m.AuthorUserID, Body: m.Body, CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}

// --- GPS ---

type GpsRepo struct{ db *gorm.DB }

func NewGpsRepo(db *gorm.DB) *GpsRepo { return &GpsRepo{db: db} }

func toGpsPoint(m GpsTrackModel) domain.GpsPoint {
	return domain.GpsPoint{
		ID: m.ID, TenantID: m.TenantID, AgentID: m.AgentID, VisitID: m.VisitID,
		Lat: m.Lat, Lng: m.Lng, Accuracy: m.Accuracy, RecordedAt: m.RecordedAt, CreatedAt: m.CreatedAt,
	}
}

func (r *GpsRepo) AddBatch(ctx context.Context, points []domain.GpsPoint) error {
	if len(points) == 0 {
		return nil
	}
	now := time.Now().UTC()
	rows := make([]GpsTrackModel, 0, len(points))
	for i := range points {
		p := &points[i]
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}
		if p.CreatedAt.IsZero() {
			p.CreatedAt = now
		}
		rows = append(rows, GpsTrackModel{
			ID: p.ID, TenantID: p.TenantID, AgentID: p.AgentID, VisitID: p.VisitID,
			Lat: p.Lat, Lng: p.Lng, Accuracy: p.Accuracy, RecordedAt: p.RecordedAt, CreatedAt: p.CreatedAt,
		})
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *GpsRepo) ListByAgent(ctx context.Context, tenantID, agentID uuid.UUID, from, to time.Time) ([]domain.GpsPoint, error) {
	var rows []GpsTrackModel
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND agent_id = ?", tenantID, agentID)
	if !from.IsZero() {
		q = q.Where("recorded_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("recorded_at <= ?", to)
	}
	if err := q.Order("recorded_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.GpsPoint, 0, len(rows))
	for _, m := range rows {
		out = append(out, toGpsPoint(m))
	}
	return out, nil
}

func (r *GpsRepo) LatestByAgent(ctx context.Context, tenantID, agentID uuid.UUID) (*domain.GpsPoint, error) {
	var m GpsTrackModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Order("recorded_at DESC").First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	p := toGpsPoint(m)
	return &p, nil
}
