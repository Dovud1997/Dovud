package application

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/syncport"
	"github.com/google/uuid"
)

type Service struct {
	agents domain.AgentRepository
	routes domain.RouteRepository
	visits domain.VisitRepository
	gps    domain.GpsRepository
	sync   syncport.ChangeRecorder
}

func NewService(
	agents domain.AgentRepository,
	routes domain.RouteRepository,
	visits domain.VisitRepository,
	gps domain.GpsRepository,
) *Service {
	return &Service{agents: agents, routes: routes, visits: visits, gps: gps}
}

func (s *Service) WithSync(rec syncport.ChangeRecorder) *Service {
	s.sync = rec
	return s
}

func (s *Service) recordVisit(ctx context.Context, tenantID uuid.UUID, dto *VisitDTO) {
	if s.sync == nil || dto == nil || !syncport.ShouldFanout(ctx) {
		return
	}
	_ = s.sync.RecordChange(ctx, tenantID, "visit", dto.ID.String(), dto.Version, false, dto)
}

// --- Agents ---

type AgentDTO struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	BranchID     uuid.UUID  `json:"branch_id"`
	EmployeeCode string     `json:"employee_code"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	Status       string     `json:"status"`
}

type AgentInput struct {
	UserID       uuid.UUID  `json:"user_id"`
	BranchID     uuid.UUID  `json:"branch_id"`
	EmployeeCode string     `json:"employee_code"`
	ManagerID    *uuid.UUID `json:"manager_id"`
	Status       string     `json:"status"`
}

func toAgentDTO(a domain.SalesAgent) AgentDTO {
	return AgentDTO{
		ID: a.ID, UserID: a.UserID, BranchID: a.BranchID,
		EmployeeCode: a.EmployeeCode, ManagerID: a.ManagerID, Status: a.Status,
	}
}

func (s *Service) ListAgents(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]AgentDTO, int64, error) {
	rows, total, err := s.agents.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]AgentDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAgentDTO(r))
	}
	return out, total, nil
}

func (s *Service) GetAgent(ctx context.Context, tenantID, id uuid.UUID) (*AgentDTO, error) {
	a, err := s.agents.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toAgentDTO(*a)
	return &dto, nil
}

func (s *Service) CreateAgent(ctx context.Context, tenantID uuid.UUID, in AgentInput) (*AgentDTO, error) {
	code := strings.TrimSpace(in.EmployeeCode)
	if in.UserID == uuid.Nil || in.BranchID == uuid.Nil || code == "" {
		return nil, apperrors.ErrValidation
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	a := &domain.SalesAgent{
		TenantID: tenantID, UserID: in.UserID, BranchID: in.BranchID,
		EmployeeCode: code, ManagerID: in.ManagerID, Status: status,
	}
	if err := s.agents.Create(ctx, a); err != nil {
		return nil, err
	}
	dto := toAgentDTO(*a)
	return &dto, nil
}

func (s *Service) UpdateAgent(ctx context.Context, tenantID, id uuid.UUID, in AgentInput) (*AgentDTO, error) {
	a, err := s.agents.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.UserID != uuid.Nil {
		a.UserID = in.UserID
	}
	if in.BranchID != uuid.Nil {
		a.BranchID = in.BranchID
	}
	if code := strings.TrimSpace(in.EmployeeCode); code != "" {
		a.EmployeeCode = code
	}
	if in.ManagerID != nil {
		a.ManagerID = in.ManagerID
	}
	if in.Status != "" {
		a.Status = in.Status
	}
	if err := s.agents.Update(ctx, a); err != nil {
		return nil, err
	}
	dto := toAgentDTO(*a)
	return &dto, nil
}

func (s *Service) DeleteAgent(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.agents.SoftDelete(ctx, tenantID, id)
}

// --- Routes ---

type RouteStopDTO struct {
	ID             uuid.UUID  `json:"id"`
	CustomerID     uuid.UUID  `json:"customer_id"`
	Sequence       int        `json:"sequence"`
	PlannedArrival *time.Time `json:"planned_arrival,omitempty"`
	Status         string     `json:"status"`
}

type RouteDTO struct {
	ID      uuid.UUID      `json:"id"`
	AgentID uuid.UUID      `json:"agent_id"`
	Date    string         `json:"date"`
	Name    string         `json:"name"`
	Status  string         `json:"status"`
	Version int64          `json:"version"`
	Stops   []RouteStopDTO `json:"stops,omitempty"`
}

type RouteInput struct {
	AgentID uuid.UUID `json:"agent_id"`
	Date    string    `json:"date"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
}

type RouteStopInput struct {
	CustomerID     uuid.UUID  `json:"customer_id"`
	Sequence       int        `json:"sequence"`
	PlannedArrival *time.Time `json:"planned_arrival"`
	Status         string     `json:"status"`
}

func toRouteStopDTO(s domain.RouteStop) RouteStopDTO {
	return RouteStopDTO{
		ID: s.ID, CustomerID: s.CustomerID, Sequence: s.Sequence,
		PlannedArrival: s.PlannedArrival, Status: s.Status,
	}
}

func toRouteDTO(r domain.Route, stops []domain.RouteStop) RouteDTO {
	dto := RouteDTO{
		ID: r.ID, AgentID: r.AgentID, Date: r.Date.Format("2006-01-02"),
		Name: r.Name, Status: r.Status, Version: r.Version,
	}
	if stops != nil {
		dto.Stops = make([]RouteStopDTO, 0, len(stops))
		for _, st := range stops {
			dto.Stops = append(dto.Stops, toRouteStopDTO(st))
		}
	}
	return dto
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, apperrors.ErrValidation
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, apperrors.ErrValidation
	}
	return t, nil
}

func validRouteStatus(st string) bool {
	switch st {
	case "planned", "in_progress", "done", "cancelled":
		return true
	default:
		return false
	}
}

func (s *Service) ListRoutes(ctx context.Context, tenantID uuid.UUID, agentID *uuid.UUID, date *time.Time, page, perPage int) ([]RouteDTO, int64, error) {
	rows, total, err := s.routes.List(ctx, tenantID, agentID, date, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]RouteDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRouteDTO(r, nil))
	}
	return out, total, nil
}

func (s *Service) GetRoute(ctx context.Context, tenantID, id uuid.UUID) (*RouteDTO, error) {
	rt, err := s.routes.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	stops, err := s.routes.ListStops(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toRouteDTO(*rt, stops)
	return &dto, nil
}

func (s *Service) CreateRoute(ctx context.Context, tenantID uuid.UUID, in RouteInput) (*RouteDTO, error) {
	name := strings.TrimSpace(in.Name)
	if in.AgentID == uuid.Nil || name == "" {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.agents.FindByID(ctx, tenantID, in.AgentID); err != nil {
		return nil, err
	}
	date, err := parseDate(in.Date)
	if err != nil {
		return nil, err
	}
	status := in.Status
	if status == "" {
		status = "planned"
	}
	if !validRouteStatus(status) {
		return nil, apperrors.ErrValidation
	}
	rt := &domain.Route{
		TenantID: tenantID, AgentID: in.AgentID, Date: date, Name: name, Status: status,
	}
	if err := s.routes.Create(ctx, rt); err != nil {
		return nil, err
	}
	dto := toRouteDTO(*rt, []domain.RouteStop{})
	return &dto, nil
}

func (s *Service) UpdateRoute(ctx context.Context, tenantID, id uuid.UUID, in RouteInput) (*RouteDTO, error) {
	rt, err := s.routes.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.AgentID != uuid.Nil {
		if _, err := s.agents.FindByID(ctx, tenantID, in.AgentID); err != nil {
			return nil, err
		}
		rt.AgentID = in.AgentID
	}
	if strings.TrimSpace(in.Date) != "" {
		date, err := parseDate(in.Date)
		if err != nil {
			return nil, err
		}
		rt.Date = date
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		rt.Name = name
	}
	if in.Status != "" {
		if !validRouteStatus(in.Status) {
			return nil, apperrors.ErrValidation
		}
		rt.Status = in.Status
	}
	if err := s.routes.Update(ctx, rt); err != nil {
		return nil, err
	}
	stops, err := s.routes.ListStops(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toRouteDTO(*rt, stops)
	return &dto, nil
}

func (s *Service) DeleteRoute(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.routes.SoftDelete(ctx, tenantID, id)
}

func (s *Service) SetStops(ctx context.Context, tenantID, routeID uuid.UUID, inputs []RouteStopInput) (*RouteDTO, error) {
	rt, err := s.routes.FindByID(ctx, tenantID, routeID)
	if err != nil {
		return nil, err
	}
	stops := make([]domain.RouteStop, 0, len(inputs))
	for _, in := range inputs {
		if in.CustomerID == uuid.Nil {
			return nil, apperrors.ErrValidation
		}
		status := in.Status
		if status == "" {
			status = "pending"
		}
		if status != "pending" && status != "visited" && status != "skipped" {
			return nil, apperrors.ErrValidation
		}
		stops = append(stops, domain.RouteStop{
			RouteID: routeID, CustomerID: in.CustomerID, Sequence: in.Sequence,
			PlannedArrival: in.PlannedArrival, Status: status,
		})
	}
	if err := s.routes.ReplaceStops(ctx, routeID, stops); err != nil {
		return nil, err
	}
	listed, err := s.routes.ListStops(ctx, routeID)
	if err != nil {
		return nil, err
	}
	dto := toRouteDTO(*rt, listed)
	return &dto, nil
}

// --- Visits ---

type VisitPhotoDTO struct {
	ID        uuid.UUID `json:"id"`
	FileURL   string    `json:"file_url"`
	Caption   *string   `json:"caption,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type VisitCommentDTO struct {
	ID           uuid.UUID `json:"id"`
	AuthorUserID uuid.UUID `json:"author_user_id"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}

type VisitDTO struct {
	ID          uuid.UUID         `json:"id"`
	AgentID     uuid.UUID         `json:"agent_id"`
	CustomerID  uuid.UUID         `json:"customer_id"`
	RouteStopID *uuid.UUID        `json:"route_stop_id,omitempty"`
	StartedAt   time.Time         `json:"started_at"`
	EndedAt     *time.Time        `json:"ended_at,omitempty"`
	CheckinLat  *float64          `json:"checkin_lat,omitempty"`
	CheckinLng  *float64          `json:"checkin_lng,omitempty"`
	CheckoutLat *float64          `json:"checkout_lat,omitempty"`
	CheckoutLng *float64          `json:"checkout_lng,omitempty"`
	Result      string            `json:"result"`
	Notes       *string           `json:"notes,omitempty"`
	Version     int64             `json:"version"`
	Photos      []VisitPhotoDTO   `json:"photos,omitempty"`
	Comments    []VisitCommentDTO `json:"comments,omitempty"`
}

type CheckInInput struct {
	AgentID     uuid.UUID  `json:"agent_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	RouteStopID *uuid.UUID `json:"route_stop_id"`
	Lat         *float64   `json:"lat"`
	Lng         *float64   `json:"lng"`
	Notes       *string    `json:"notes"`
}

type CheckOutInput struct {
	Lat    *float64 `json:"lat"`
	Lng    *float64 `json:"lng"`
	Result string   `json:"result"`
	Notes  *string  `json:"notes"`
}

type PhotoInput struct {
	FileURL string  `json:"file_url"`
	Caption *string `json:"caption"`
}

type CommentInput struct {
	Body string `json:"body"`
}

func toVisitDTO(v domain.Visit, photos []domain.VisitPhoto, comments []domain.VisitComment) VisitDTO {
	dto := VisitDTO{
		ID: v.ID, AgentID: v.AgentID, CustomerID: v.CustomerID, RouteStopID: v.RouteStopID,
		StartedAt: v.StartedAt, EndedAt: v.EndedAt,
		CheckinLat: v.CheckinLat, CheckinLng: v.CheckinLng, CheckoutLat: v.CheckoutLat, CheckoutLng: v.CheckoutLng,
		Result: v.Result, Notes: v.Notes, Version: v.Version,
	}
	if photos != nil {
		dto.Photos = make([]VisitPhotoDTO, 0, len(photos))
		for _, p := range photos {
			dto.Photos = append(dto.Photos, VisitPhotoDTO{
				ID: p.ID, FileURL: p.FileURL, Caption: p.Caption, CreatedAt: p.CreatedAt,
			})
		}
	}
	if comments != nil {
		dto.Comments = make([]VisitCommentDTO, 0, len(comments))
		for _, c := range comments {
			dto.Comments = append(dto.Comments, VisitCommentDTO{
				ID: c.ID, AuthorUserID: c.AuthorUserID, Body: c.Body, CreatedAt: c.CreatedAt,
			})
		}
	}
	return dto
}

func validVisitResult(r string) bool {
	switch r {
	case "success", "no_order", "closed", "other":
		return true
	default:
		return false
	}
}

func (s *Service) ListVisits(ctx context.Context, tenantID uuid.UUID, agentID, customerID *uuid.UUID, page, perPage int) ([]VisitDTO, int64, error) {
	rows, total, err := s.visits.List(ctx, tenantID, agentID, customerID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]VisitDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toVisitDTO(r, nil, nil))
	}
	return out, total, nil
}

func (s *Service) GetVisit(ctx context.Context, tenantID, id uuid.UUID) (*VisitDTO, error) {
	v, err := s.visits.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	photos, err := s.visits.ListPhotos(ctx, id)
	if err != nil {
		return nil, err
	}
	comments, err := s.visits.ListComments(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toVisitDTO(*v, photos, comments)
	return &dto, nil
}

func (s *Service) CheckIn(ctx context.Context, tenantID uuid.UUID, in CheckInInput) (*VisitDTO, error) {
	if in.AgentID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.agents.FindByID(ctx, tenantID, in.AgentID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	v := &domain.Visit{
		TenantID: tenantID, AgentID: in.AgentID, CustomerID: in.CustomerID,
		RouteStopID: in.RouteStopID, StartedAt: now,
		CheckinLat: in.Lat, CheckinLng: in.Lng, Notes: in.Notes, Result: "",
	}
	if err := s.visits.Create(ctx, v); err != nil {
		return nil, err
	}
	dto := toVisitDTO(*v, []domain.VisitPhoto{}, []domain.VisitComment{})
	s.recordVisit(ctx, tenantID, &dto)
	return &dto, nil
}

func (s *Service) CheckOut(ctx context.Context, tenantID, visitID uuid.UUID, in CheckOutInput) (*VisitDTO, error) {
	v, err := s.visits.FindByID(ctx, tenantID, visitID)
	if err != nil {
		return nil, err
	}
	result := strings.TrimSpace(in.Result)
	if result == "" || !validVisitResult(result) {
		return nil, apperrors.ErrValidation
	}
	now := time.Now().UTC()
	v.EndedAt = &now
	v.CheckoutLat = in.Lat
	v.CheckoutLng = in.Lng
	v.Result = result
	if in.Notes != nil {
		v.Notes = in.Notes
	}
	if err := s.visits.Update(ctx, v); err != nil {
		return nil, err
	}
	dto := toVisitDTO(*v, nil, nil)
	s.recordVisit(ctx, tenantID, &dto)
	return &dto, nil
}

func (s *Service) AddPhoto(ctx context.Context, tenantID, visitID uuid.UUID, in PhotoInput) (*VisitPhotoDTO, error) {
	if _, err := s.visits.FindByID(ctx, tenantID, visitID); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(in.FileURL)
	if url == "" {
		return nil, apperrors.ErrValidation
	}
	p := &domain.VisitPhoto{VisitID: visitID, FileURL: url, Caption: in.Caption}
	if err := s.visits.AddPhoto(ctx, p); err != nil {
		return nil, err
	}
	return &VisitPhotoDTO{ID: p.ID, FileURL: p.FileURL, Caption: p.Caption, CreatedAt: p.CreatedAt}, nil
}

func (s *Service) AddComment(ctx context.Context, tenantID, visitID, authorUserID uuid.UUID, in CommentInput) (*VisitCommentDTO, error) {
	if _, err := s.visits.FindByID(ctx, tenantID, visitID); err != nil {
		return nil, err
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, apperrors.ErrValidation
	}
	c := &domain.VisitComment{VisitID: visitID, AuthorUserID: authorUserID, Body: body}
	if err := s.visits.AddComment(ctx, c); err != nil {
		return nil, err
	}
	return &VisitCommentDTO{ID: c.ID, AuthorUserID: c.AuthorUserID, Body: c.Body, CreatedAt: c.CreatedAt}, nil
}

func (s *Service) ListComments(ctx context.Context, tenantID, visitID uuid.UUID) ([]VisitCommentDTO, error) {
	if _, err := s.visits.FindByID(ctx, tenantID, visitID); err != nil {
		return nil, err
	}
	rows, err := s.visits.ListComments(ctx, visitID)
	if err != nil {
		return nil, err
	}
	out := make([]VisitCommentDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, VisitCommentDTO{ID: r.ID, AuthorUserID: r.AuthorUserID, Body: r.Body, CreatedAt: r.CreatedAt})
	}
	return out, nil
}

// --- GPS ---

type GpsPointDTO struct {
	ID         uuid.UUID  `json:"id"`
	AgentID    uuid.UUID  `json:"agent_id"`
	VisitID    *uuid.UUID `json:"visit_id,omitempty"`
	Lat        float64    `json:"lat"`
	Lng        float64    `json:"lng"`
	Accuracy   *float64   `json:"accuracy,omitempty"`
	RecordedAt time.Time  `json:"recorded_at"`
}

type GpsPointInput struct {
	AgentID    uuid.UUID  `json:"agent_id"`
	VisitID    *uuid.UUID `json:"visit_id"`
	Lat        float64    `json:"lat"`
	Lng        float64    `json:"lng"`
	Accuracy   *float64   `json:"accuracy"`
	RecordedAt time.Time  `json:"recorded_at"`
}

func toGpsDTO(p domain.GpsPoint) GpsPointDTO {
	return GpsPointDTO{
		ID: p.ID, AgentID: p.AgentID, VisitID: p.VisitID,
		Lat: p.Lat, Lng: p.Lng, Accuracy: p.Accuracy, RecordedAt: p.RecordedAt,
	}
}

func (s *Service) UploadPoints(ctx context.Context, tenantID uuid.UUID, inputs []GpsPointInput) ([]GpsPointDTO, error) {
	if len(inputs) == 0 {
		return nil, apperrors.ErrValidation
	}
	seen := map[uuid.UUID]struct{}{}
	points := make([]domain.GpsPoint, 0, len(inputs))
	for _, in := range inputs {
		if in.AgentID == uuid.Nil || in.RecordedAt.IsZero() {
			return nil, apperrors.ErrValidation
		}
		if _, ok := seen[in.AgentID]; !ok {
			if _, err := s.agents.FindByID(ctx, tenantID, in.AgentID); err != nil {
				return nil, err
			}
			seen[in.AgentID] = struct{}{}
		}
		points = append(points, domain.GpsPoint{
			TenantID: tenantID, AgentID: in.AgentID, VisitID: in.VisitID,
			Lat: in.Lat, Lng: in.Lng, Accuracy: in.Accuracy, RecordedAt: in.RecordedAt.UTC(),
		})
	}
	if err := s.gps.AddBatch(ctx, points); err != nil {
		return nil, err
	}
	out := make([]GpsPointDTO, 0, len(points))
	for _, p := range points {
		out = append(out, toGpsDTO(p))
	}
	return out, nil
}

func (s *Service) LivePosition(ctx context.Context, tenantID, agentID uuid.UUID) (*GpsPointDTO, error) {
	if _, err := s.agents.FindByID(ctx, tenantID, agentID); err != nil {
		return nil, err
	}
	p, err := s.gps.LatestByAgent(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}
	dto := toGpsDTO(*p)
	return &dto, nil
}

func (s *Service) TrackHistory(ctx context.Context, tenantID, agentID uuid.UUID, from, to time.Time) ([]GpsPointDTO, error) {
	if _, err := s.agents.FindByID(ctx, tenantID, agentID); err != nil {
		return nil, err
	}
	rows, err := s.gps.ListByAgent(ctx, tenantID, agentID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]GpsPointDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toGpsDTO(r))
	}
	return out, nil
}
