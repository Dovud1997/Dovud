package application

import (
	"context"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/organization/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	companies   domain.CompanyRepository
	branches    domain.BranchRepository
	warehouses  domain.WarehouseRepository
}

func NewService(c domain.CompanyRepository, b domain.BranchRepository, w domain.WarehouseRepository) *Service {
	return &Service{companies: c, branches: b, warehouses: w}
}

type CompanyDTO struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Name   string    `json:"name"`
	Inn    *string   `json:"inn,omitempty"`
	Status string    `json:"status"`
}

type CompanyInput struct {
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Inn    *string `json:"inn"`
	Status string  `json:"status"`
}

func (s *Service) ListCompanies(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]CompanyDTO, int64, error) {
	rows, total, err := s.companies.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]CompanyDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, CompanyDTO{ID: r.ID, Code: r.Code, Name: r.Name, Inn: r.Inn, Status: r.Status})
	}
	return out, total, nil
}

func (s *Service) CreateCompany(ctx context.Context, tenantID uuid.UUID, in CompanyInput) (*CompanyDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	c := &domain.Company{TenantID: tenantID, Code: code, Name: name, Inn: in.Inn, Status: status}
	if err := s.companies.Create(ctx, c); err != nil {
		return nil, err
	}
	return &CompanyDTO{ID: c.ID, Code: c.Code, Name: c.Name, Inn: c.Inn, Status: c.Status}, nil
}

func (s *Service) UpdateCompany(ctx context.Context, tenantID, id uuid.UUID, in CompanyInput) (*CompanyDTO, error) {
	c, err := s.companies.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		c.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		c.Name = strings.TrimSpace(in.Name)
	}
	if in.Inn != nil {
		c.Inn = in.Inn
	}
	if in.Status != "" {
		c.Status = in.Status
	}
	if err := s.companies.Update(ctx, c); err != nil {
		return nil, err
	}
	return &CompanyDTO{ID: c.ID, Code: c.Code, Name: c.Name, Inn: c.Inn, Status: c.Status}, nil
}

func (s *Service) DeleteCompany(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.companies.SoftDelete(ctx, tenantID, id)
}

type BranchDTO struct {
	ID        uuid.UUID  `json:"id"`
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Address   *string    `json:"address,omitempty"`
	Lat       *float64   `json:"lat,omitempty"`
	Lng       *float64   `json:"lng,omitempty"`
	Status    string     `json:"status"`
}

type BranchInput struct {
	CompanyID *uuid.UUID `json:"company_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Address   *string    `json:"address"`
	Lat       *float64   `json:"lat"`
	Lng       *float64   `json:"lng"`
	Status    string     `json:"status"`
}

func (s *Service) ListBranches(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]BranchDTO, int64, error) {
	rows, total, err := s.branches.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]BranchDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, BranchDTO{ID: r.ID, CompanyID: r.CompanyID, Code: r.Code, Name: r.Name, Address: r.Address, Lat: r.Lat, Lng: r.Lng, Status: r.Status})
	}
	return out, total, nil
}

func (s *Service) CreateBranch(ctx context.Context, tenantID uuid.UUID, in BranchInput) (*BranchDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	if in.CompanyID != nil {
		if _, err := s.companies.FindByID(ctx, tenantID, *in.CompanyID); err != nil {
			return nil, apperrors.ErrValidation
		}
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	b := &domain.Branch{TenantID: tenantID, CompanyID: in.CompanyID, Code: code, Name: name, Address: in.Address, Lat: in.Lat, Lng: in.Lng, Status: status}
	if err := s.branches.Create(ctx, b); err != nil {
		return nil, err
	}
	return &BranchDTO{ID: b.ID, CompanyID: b.CompanyID, Code: b.Code, Name: b.Name, Address: b.Address, Lat: b.Lat, Lng: b.Lng, Status: b.Status}, nil
}

func (s *Service) UpdateBranch(ctx context.Context, tenantID, id uuid.UUID, in BranchInput) (*BranchDTO, error) {
	b, err := s.branches.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		b.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		b.Name = strings.TrimSpace(in.Name)
	}
	if in.CompanyID != nil {
		b.CompanyID = in.CompanyID
	}
	if in.Address != nil {
		b.Address = in.Address
	}
	if in.Lat != nil {
		b.Lat = in.Lat
	}
	if in.Lng != nil {
		b.Lng = in.Lng
	}
	if in.Status != "" {
		b.Status = in.Status
	}
	if err := s.branches.Update(ctx, b); err != nil {
		return nil, err
	}
	return &BranchDTO{ID: b.ID, CompanyID: b.CompanyID, Code: b.Code, Name: b.Name, Address: b.Address, Lat: b.Lat, Lng: b.Lng, Status: b.Status}, nil
}

func (s *Service) DeleteBranch(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.branches.SoftDelete(ctx, tenantID, id)
}

type WarehouseDTO struct {
	ID       uuid.UUID `json:"id"`
	BranchID uuid.UUID `json:"branch_id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Status   string    `json:"status"`
}

type WarehouseInput struct {
	BranchID uuid.UUID `json:"branch_id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Status   string    `json:"status"`
}

type StockDTO struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	QtyOnHand   float64   `json:"qty_on_hand"`
	QtyReserved float64   `json:"qty_reserved"`
}

type StockInput struct {
	QtyOnHand   float64 `json:"qty_on_hand"`
	QtyReserved float64 `json:"qty_reserved"`
}

func (s *Service) ListWarehouses(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID, page, perPage int) ([]WarehouseDTO, int64, error) {
	rows, total, err := s.warehouses.List(ctx, tenantID, branchID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]WarehouseDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, WarehouseDTO{ID: r.ID, BranchID: r.BranchID, Code: r.Code, Name: r.Name, Type: r.Type, Status: r.Status})
	}
	return out, total, nil
}

func (s *Service) CreateWarehouse(ctx context.Context, tenantID uuid.UUID, in WarehouseInput) (*WarehouseDTO, error) {
	if in.BranchID == uuid.Nil || strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.Name) == "" {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.branches.FindByID(ctx, tenantID, in.BranchID); err != nil {
		return nil, apperrors.ErrValidation
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	w := &domain.Warehouse{TenantID: tenantID, BranchID: in.BranchID, Code: in.Code, Name: strings.TrimSpace(in.Name), Type: in.Type, Status: status}
	if err := s.warehouses.Create(ctx, w); err != nil {
		return nil, err
	}
	return &WarehouseDTO{ID: w.ID, BranchID: w.BranchID, Code: w.Code, Name: w.Name, Type: w.Type, Status: w.Status}, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, tenantID, id uuid.UUID, in WarehouseInput) (*WarehouseDTO, error) {
	w, err := s.warehouses.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.BranchID != uuid.Nil {
		w.BranchID = in.BranchID
	}
	if strings.TrimSpace(in.Code) != "" {
		w.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		w.Name = strings.TrimSpace(in.Name)
	}
	if in.Type != "" {
		w.Type = in.Type
	}
	if in.Status != "" {
		w.Status = in.Status
	}
	if err := s.warehouses.Update(ctx, w); err != nil {
		return nil, err
	}
	return &WarehouseDTO{ID: w.ID, BranchID: w.BranchID, Code: w.Code, Name: w.Name, Type: w.Type, Status: w.Status}, nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.warehouses.SoftDelete(ctx, tenantID, id)
}

func (s *Service) ListStocks(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]StockDTO, error) {
	if _, err := s.warehouses.FindByID(ctx, tenantID, warehouseID); err != nil {
		return nil, err
	}
	rows, err := s.warehouses.ListStocks(ctx, tenantID, warehouseID)
	if err != nil {
		return nil, err
	}
	out := make([]StockDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, StockDTO{ID: r.ID, ProductID: r.ProductID, QtyOnHand: r.QtyOnHand, QtyReserved: r.QtyReserved})
	}
	return out, nil
}

func (s *Service) UpsertStock(ctx context.Context, tenantID, warehouseID, productID uuid.UUID, in StockInput) (*StockDTO, error) {
	if _, err := s.warehouses.FindByID(ctx, tenantID, warehouseID); err != nil {
		return nil, err
	}
	stock := &domain.WarehouseStock{TenantID: tenantID, WarehouseID: warehouseID, ProductID: productID, QtyOnHand: in.QtyOnHand, QtyReserved: in.QtyReserved}
	if err := s.warehouses.UpsertStock(ctx, stock); err != nil {
		return nil, err
	}
	return &StockDTO{ID: stock.ID, ProductID: stock.ProductID, QtyOnHand: stock.QtyOnHand, QtyReserved: stock.QtyReserved}, nil
}
