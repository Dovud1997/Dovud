package application

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/catalog/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	manufacturers domain.ManufacturerRepository
	categories    domain.CategoryRepository
	products      domain.ProductRepository
	prices        domain.PriceRepository
	promotions    domain.PromotionRepository
}

func NewService(
	m domain.ManufacturerRepository,
	c domain.CategoryRepository,
	p domain.ProductRepository,
	pr domain.PriceRepository,
	promo domain.PromotionRepository,
) *Service {
	return &Service{manufacturers: m, categories: c, products: p, prices: pr, promotions: promo}
}

type ManufacturerDTO struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

type ManufacturerInput struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *Service) ListManufacturers(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]ManufacturerDTO, int64, error) {
	rows, total, err := s.manufacturers.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ManufacturerDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ManufacturerDTO{ID: r.ID, Code: r.Code, Name: r.Name, Status: r.Status})
	}
	return out, total, nil
}

func (s *Service) CreateManufacturer(ctx context.Context, tenantID uuid.UUID, in ManufacturerInput) (*ManufacturerDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	m := &domain.Manufacturer{TenantID: tenantID, Code: code, Name: name, Status: status}
	if err := s.manufacturers.Create(ctx, m); err != nil {
		return nil, err
	}
	return &ManufacturerDTO{ID: m.ID, Code: m.Code, Name: m.Name, Status: m.Status}, nil
}

func (s *Service) UpdateManufacturer(ctx context.Context, tenantID, id uuid.UUID, in ManufacturerInput) (*ManufacturerDTO, error) {
	m, err := s.manufacturers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		m.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		m.Name = strings.TrimSpace(in.Name)
	}
	if in.Status != "" {
		m.Status = in.Status
	}
	if err := s.manufacturers.Update(ctx, m); err != nil {
		return nil, err
	}
	return &ManufacturerDTO{ID: m.ID, Code: m.Code, Name: m.Name, Status: m.Status}, nil
}

func (s *Service) DeleteManufacturer(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.manufacturers.SoftDelete(ctx, tenantID, id)
}

type CategoryDTO struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sort_order"`
}

type CategoryInput struct {
	ParentID  *uuid.UUID `json:"parent_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sort_order"`
}

func (s *Service) ListCategories(ctx context.Context, tenantID uuid.UUID) ([]CategoryDTO, error) {
	rows, err := s.categories.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, CategoryDTO{ID: r.ID, ParentID: r.ParentID, Code: r.Code, Name: r.Name, SortOrder: r.SortOrder})
	}
	return out, nil
}

func (s *Service) CreateCategory(ctx context.Context, tenantID uuid.UUID, in CategoryInput) (*CategoryDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	c := &domain.Category{TenantID: tenantID, ParentID: in.ParentID, Code: code, Name: name, SortOrder: in.SortOrder}
	if err := s.categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return &CategoryDTO{ID: c.ID, ParentID: c.ParentID, Code: c.Code, Name: c.Name, SortOrder: c.SortOrder}, nil
}

func (s *Service) UpdateCategory(ctx context.Context, tenantID, id uuid.UUID, in CategoryInput) (*CategoryDTO, error) {
	c, err := s.categories.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		c.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		c.Name = strings.TrimSpace(in.Name)
	}
	c.ParentID = in.ParentID
	c.SortOrder = in.SortOrder
	if err := s.categories.Update(ctx, c); err != nil {
		return nil, err
	}
	return &CategoryDTO{ID: c.ID, ParentID: c.ParentID, Code: c.Code, Name: c.Name, SortOrder: c.SortOrder}, nil
}

func (s *Service) DeleteCategory(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.categories.SoftDelete(ctx, tenantID, id)
}

type ProductDTO struct {
	ID             uuid.UUID  `json:"id"`
	SKU            string     `json:"sku"`
	Barcode        *string    `json:"barcode,omitempty"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	ManufacturerID *uuid.UUID `json:"manufacturer_id,omitempty"`
	Unit           string     `json:"unit"`
	VATRate        float64    `json:"vat_rate"`
	IsActive       bool       `json:"is_active"`
	Version        int64      `json:"version"`
}

type ProductInput struct {
	SKU            string     `json:"sku"`
	Barcode        *string    `json:"barcode"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	CategoryID     *uuid.UUID `json:"category_id"`
	ManufacturerID *uuid.UUID `json:"manufacturer_id"`
	Unit           string     `json:"unit"`
	VATRate        float64    `json:"vat_rate"`
	IsActive       *bool      `json:"is_active"`
}

func toProductDTO(p domain.Product) ProductDTO {
	return ProductDTO{ID: p.ID, SKU: p.SKU, Barcode: p.Barcode, Name: p.Name, Description: p.Description, CategoryID: p.CategoryID, ManufacturerID: p.ManufacturerID, Unit: p.Unit, VATRate: p.VATRate, IsActive: p.IsActive, Version: p.Version}
}

func (s *Service) ListProducts(ctx context.Context, tenantID uuid.UUID, q string, page, perPage int) ([]ProductDTO, int64, error) {
	rows, total, err := s.products.List(ctx, tenantID, q, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ProductDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toProductDTO(r))
	}
	return out, total, nil
}

func (s *Service) GetProduct(ctx context.Context, tenantID, id uuid.UUID) (*ProductDTO, error) {
	p, err := s.products.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toProductDTO(*p)
	return &dto, nil
}

func (s *Service) CreateProduct(ctx context.Context, tenantID uuid.UUID, in ProductInput) (*ProductDTO, error) {
	sku := strings.ToUpper(strings.TrimSpace(in.SKU))
	name := strings.TrimSpace(in.Name)
	if sku == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := &domain.Product{TenantID: tenantID, SKU: sku, Barcode: in.Barcode, Name: name, Description: in.Description, CategoryID: in.CategoryID, ManufacturerID: in.ManufacturerID, Unit: in.Unit, VATRate: in.VATRate, IsActive: active}
	if err := s.products.Create(ctx, p); err != nil {
		return nil, err
	}
	dto := toProductDTO(*p)
	return &dto, nil
}

func (s *Service) UpdateProduct(ctx context.Context, tenantID, id uuid.UUID, in ProductInput) (*ProductDTO, error) {
	p, err := s.products.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.SKU) != "" {
		p.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	}
	if strings.TrimSpace(in.Name) != "" {
		p.Name = strings.TrimSpace(in.Name)
	}
	if in.Barcode != nil {
		p.Barcode = in.Barcode
	}
	if in.Description != nil {
		p.Description = in.Description
	}
	p.CategoryID = in.CategoryID
	p.ManufacturerID = in.ManufacturerID
	if in.Unit != "" {
		p.Unit = in.Unit
	}
	p.VATRate = in.VATRate
	if in.IsActive != nil {
		p.IsActive = *in.IsActive
	}
	if err := s.products.Update(ctx, p); err != nil {
		return nil, err
	}
	dto := toProductDTO(*p)
	return &dto, nil
}

func (s *Service) DeleteProduct(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.products.SoftDelete(ctx, tenantID, id)
}

type PriceListDTO struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	IsDefault bool      `json:"is_default"`
}

type PriceListInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	IsDefault bool   `json:"is_default"`
}

type PriceDTO struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	Amount    float64    `json:"amount"`
	Currency  string     `json:"currency"`
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
	Version   int64      `json:"version"`
}

type PriceInput struct {
	ProductID uuid.UUID  `json:"product_id"`
	Amount    float64    `json:"amount"`
	Currency  string     `json:"currency"`
	ValidFrom *time.Time `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to"`
}

func (s *Service) ListPriceLists(ctx context.Context, tenantID uuid.UUID) ([]PriceListDTO, error) {
	rows, err := s.prices.ListPriceLists(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]PriceListDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, PriceListDTO{ID: r.ID, Code: r.Code, Name: r.Name, Currency: r.Currency, IsDefault: r.IsDefault})
	}
	return out, nil
}

func (s *Service) CreatePriceList(ctx context.Context, tenantID uuid.UUID, in PriceListInput) (*PriceListDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if code == "" || name == "" || currency == "" {
		return nil, apperrors.ErrValidation
	}
	pl := &domain.PriceList{TenantID: tenantID, Code: code, Name: name, Currency: currency, IsDefault: in.IsDefault}
	if err := s.prices.CreatePriceList(ctx, pl); err != nil {
		return nil, err
	}
	return &PriceListDTO{ID: pl.ID, Code: pl.Code, Name: pl.Name, Currency: pl.Currency, IsDefault: pl.IsDefault}, nil
}

func (s *Service) UpdatePriceList(ctx context.Context, tenantID, id uuid.UUID, in PriceListInput) (*PriceListDTO, error) {
	pl, err := s.prices.FindPriceList(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		pl.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		pl.Name = strings.TrimSpace(in.Name)
	}
	if strings.TrimSpace(in.Currency) != "" {
		pl.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	}
	pl.IsDefault = in.IsDefault
	if err := s.prices.UpdatePriceList(ctx, pl); err != nil {
		return nil, err
	}
	return &PriceListDTO{ID: pl.ID, Code: pl.Code, Name: pl.Name, Currency: pl.Currency, IsDefault: pl.IsDefault}, nil
}

func (s *Service) DeletePriceList(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.prices.SoftDeletePriceList(ctx, tenantID, id)
}

func (s *Service) ListPrices(ctx context.Context, tenantID, priceListID uuid.UUID) ([]PriceDTO, error) {
	if _, err := s.prices.FindPriceList(ctx, tenantID, priceListID); err != nil {
		return nil, err
	}
	rows, err := s.prices.ListPrices(ctx, tenantID, priceListID)
	if err != nil {
		return nil, err
	}
	out := make([]PriceDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, PriceDTO{ID: r.ID, ProductID: r.ProductID, Amount: r.Amount, Currency: r.Currency, ValidFrom: r.ValidFrom, ValidTo: r.ValidTo, Version: r.Version})
	}
	return out, nil
}

func (s *Service) UpsertPrice(ctx context.Context, tenantID, priceListID uuid.UUID, in PriceInput) (*PriceDTO, error) {
	pl, err := s.prices.FindPriceList(ctx, tenantID, priceListID)
	if err != nil {
		return nil, err
	}
	if _, err := s.products.FindByID(ctx, tenantID, in.ProductID); err != nil {
		return nil, apperrors.ErrValidation
	}
	currency := in.Currency
	if currency == "" {
		currency = pl.Currency
	}
	price := &domain.ProductPrice{TenantID: tenantID, PriceListID: priceListID, ProductID: in.ProductID, Amount: in.Amount, Currency: currency, ValidFrom: in.ValidFrom, ValidTo: in.ValidTo}
	if err := s.prices.UpsertPrice(ctx, price); err != nil {
		return nil, err
	}
	return &PriceDTO{ID: price.ID, ProductID: price.ProductID, Amount: price.Amount, Currency: price.Currency, ValidFrom: price.ValidFrom, ValidTo: price.ValidTo, Version: price.Version}, nil
}

type PromotionDTO struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      time.Time  `json:"ends_at"`
	DiscountPct float64    `json:"discount_pct"`
	IsActive    bool       `json:"is_active"`
	Items       []PromoItemDTO `json:"items,omitempty"`
}

type PromoItemDTO struct {
	ID         uuid.UUID  `json:"id"`
	ProductID  *uuid.UUID `json:"product_id,omitempty"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
}

type PromotionInput struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	DiscountPct float64   `json:"discount_pct"`
	IsActive    *bool     `json:"is_active"`
}

type PromoItemsInput struct {
	Items []struct {
		ProductID  *uuid.UUID `json:"product_id"`
		CategoryID *uuid.UUID `json:"category_id"`
	} `json:"items"`
}

func (s *Service) ListPromotions(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]PromotionDTO, int64, error) {
	rows, total, err := s.promotions.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]PromotionDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, PromotionDTO{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, StartsAt: r.StartsAt, EndsAt: r.EndsAt, DiscountPct: r.DiscountPct, IsActive: r.IsActive})
	}
	return out, total, nil
}

func (s *Service) GetPromotion(ctx context.Context, tenantID, id uuid.UUID) (*PromotionDTO, error) {
	p, err := s.promotions.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	items, err := s.promotions.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := PromotionDTO{ID: p.ID, Code: p.Code, Name: p.Name, Description: p.Description, StartsAt: p.StartsAt, EndsAt: p.EndsAt, DiscountPct: p.DiscountPct, IsActive: p.IsActive}
	for _, it := range items {
		dto.Items = append(dto.Items, PromoItemDTO{ID: it.ID, ProductID: it.ProductID, CategoryID: it.CategoryID})
	}
	return &dto, nil
}

func (s *Service) CreatePromotion(ctx context.Context, tenantID uuid.UUID, in PromotionInput) (*PromotionDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" || in.StartsAt.IsZero() || in.EndsAt.IsZero() {
		return nil, apperrors.ErrValidation
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	p := &domain.Promotion{TenantID: tenantID, Code: code, Name: name, Description: in.Description, StartsAt: in.StartsAt, EndsAt: in.EndsAt, DiscountPct: in.DiscountPct, IsActive: active}
	if err := s.promotions.Create(ctx, p); err != nil {
		return nil, err
	}
	return &PromotionDTO{ID: p.ID, Code: p.Code, Name: p.Name, Description: p.Description, StartsAt: p.StartsAt, EndsAt: p.EndsAt, DiscountPct: p.DiscountPct, IsActive: p.IsActive}, nil
}

func (s *Service) UpdatePromotion(ctx context.Context, tenantID, id uuid.UUID, in PromotionInput) (*PromotionDTO, error) {
	p, err := s.promotions.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		p.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		p.Name = strings.TrimSpace(in.Name)
	}
	if in.Description != nil {
		p.Description = in.Description
	}
	if !in.StartsAt.IsZero() {
		p.StartsAt = in.StartsAt
	}
	if !in.EndsAt.IsZero() {
		p.EndsAt = in.EndsAt
	}
	p.DiscountPct = in.DiscountPct
	if in.IsActive != nil {
		p.IsActive = *in.IsActive
	}
	if err := s.promotions.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.GetPromotion(ctx, tenantID, id)
}

func (s *Service) DeletePromotion(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.promotions.SoftDelete(ctx, tenantID, id)
}

func (s *Service) SetPromotionItems(ctx context.Context, tenantID, id uuid.UUID, in PromoItemsInput) (*PromotionDTO, error) {
	if _, err := s.promotions.FindByID(ctx, tenantID, id); err != nil {
		return nil, err
	}
	items := make([]domain.PromotionItem, 0, len(in.Items))
	for _, it := range in.Items {
		items = append(items, domain.PromotionItem{PromotionID: id, ProductID: it.ProductID, CategoryID: it.CategoryID})
	}
	if err := s.promotions.ReplaceItems(ctx, id, items); err != nil {
		return nil, err
	}
	return s.GetPromotion(ctx, tenantID, id)
}
