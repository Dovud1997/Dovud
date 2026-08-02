package application

import (
	"context"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	customers  domain.CustomerRepository
	contacts   domain.CustomerContactRepository
	addresses  domain.CustomerAddressRepository
	categories domain.CustomerCategoryRepository
}

func NewService(
	customers domain.CustomerRepository,
	contacts domain.CustomerContactRepository,
	addresses domain.CustomerAddressRepository,
	categories domain.CustomerCategoryRepository,
) *Service {
	return &Service{customers: customers, contacts: contacts, addresses: addresses, categories: categories}
}

type CustomerDTO struct {
	ID            uuid.UUID  `json:"id"`
	BranchID      *uuid.UUID `json:"branch_id,omitempty"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Inn           *string    `json:"inn,omitempty"`
	Status        string     `json:"status"`
	CreditLimit   float64    `json:"credit_limit"`
	BalanceCached float64    `json:"balance_cached"`
	Lat           *float64   `json:"lat,omitempty"`
	Lng           *float64   `json:"lng,omitempty"`
	Address       *string    `json:"address,omitempty"`
	Version       int64      `json:"version"`
}

type CustomerInput struct {
	BranchID    *uuid.UUID `json:"branch_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Inn         *string    `json:"inn"`
	Status      string     `json:"status"`
	CreditLimit *float64   `json:"credit_limit"`
	Lat         *float64   `json:"lat"`
	Lng         *float64   `json:"lng"`
	Address     *string    `json:"address"`
}

func toCustomerDTO(c domain.Customer) CustomerDTO {
	return CustomerDTO{
		ID: c.ID, BranchID: c.BranchID, Code: c.Code, Name: c.Name, Type: c.Type, Inn: c.Inn,
		Status: c.Status, CreditLimit: c.CreditLimit, BalanceCached: c.BalanceCached,
		Lat: c.Lat, Lng: c.Lng, Address: c.Address, Version: c.Version,
	}
}

func (s *Service) ListCustomers(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]CustomerDTO, int64, error) {
	rows, total, err := s.customers.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]CustomerDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toCustomerDTO(r))
	}
	return out, total, nil
}

func (s *Service) GetCustomer(ctx context.Context, tenantID, id uuid.UUID) (*CustomerDTO, error) {
	c, err := s.customers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toCustomerDTO(*c)
	return &dto, nil
}

func (s *Service) CreateCustomer(ctx context.Context, tenantID uuid.UUID, in CustomerInput) (*CustomerDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	typ := strings.TrimSpace(in.Type)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	if typ == "" {
		typ = "outlet"
	}
	if typ != "outlet" && typ != "wholesale" {
		return nil, apperrors.ErrValidation
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	var creditLimit float64
	if in.CreditLimit != nil {
		creditLimit = *in.CreditLimit
	}
	c := &domain.Customer{
		TenantID: tenantID, BranchID: in.BranchID, Code: code, Name: name, Type: typ,
		Inn: in.Inn, Status: status, CreditLimit: creditLimit, Lat: in.Lat, Lng: in.Lng, Address: in.Address,
	}
	if err := s.customers.Create(ctx, c); err != nil {
		return nil, err
	}
	dto := toCustomerDTO(*c)
	return &dto, nil
}

func (s *Service) UpdateCustomer(ctx context.Context, tenantID, id uuid.UUID, in CustomerInput) (*CustomerDTO, error) {
	c, err := s.customers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) != "" {
		c.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	}
	if strings.TrimSpace(in.Name) != "" {
		c.Name = strings.TrimSpace(in.Name)
	}
	if typ := strings.TrimSpace(in.Type); typ != "" {
		if typ != "outlet" && typ != "wholesale" {
			return nil, apperrors.ErrValidation
		}
		c.Type = typ
	}
	if in.Inn != nil {
		c.Inn = in.Inn
	}
	if in.Status != "" {
		c.Status = in.Status
	}
	if in.BranchID != nil {
		c.BranchID = in.BranchID
	}
	if in.CreditLimit != nil {
		c.CreditLimit = *in.CreditLimit
	}
	if in.Lat != nil {
		c.Lat = in.Lat
	}
	if in.Lng != nil {
		c.Lng = in.Lng
	}
	if in.Address != nil {
		c.Address = in.Address
	}
	if err := s.customers.Update(ctx, c); err != nil {
		return nil, err
	}
	dto := toCustomerDTO(*c)
	return &dto, nil
}

func (s *Service) DeleteCustomer(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.customers.SoftDelete(ctx, tenantID, id)
}

type ContactDTO struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	Email      *string   `json:"email,omitempty"`
	Position   *string   `json:"position,omitempty"`
	IsPrimary  bool      `json:"is_primary"`
}

type ContactInput struct {
	FullName  string  `json:"full_name"`
	Phone     string  `json:"phone"`
	Email     *string `json:"email"`
	Position  *string `json:"position"`
	IsPrimary bool    `json:"is_primary"`
}

func toContactDTO(c domain.CustomerContact) ContactDTO {
	return ContactDTO{
		ID: c.ID, CustomerID: c.CustomerID, FullName: c.FullName, Phone: c.Phone,
		Email: c.Email, Position: c.Position, IsPrimary: c.IsPrimary,
	}
}

func (s *Service) ListContacts(ctx context.Context, tenantID, customerID uuid.UUID) ([]ContactDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	rows, err := s.contacts.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}
	out := make([]ContactDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toContactDTO(r))
	}
	return out, nil
}

func (s *Service) CreateContact(ctx context.Context, tenantID, customerID uuid.UUID, in ContactInput) (*ContactDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.FullName)
	phone := strings.TrimSpace(in.Phone)
	if name == "" || phone == "" {
		return nil, apperrors.ErrValidation
	}
	c := &domain.CustomerContact{
		CustomerID: customerID, FullName: name, Phone: phone,
		Email: in.Email, Position: in.Position, IsPrimary: in.IsPrimary,
	}
	if err := s.contacts.Create(ctx, c); err != nil {
		return nil, err
	}
	dto := toContactDTO(*c)
	return &dto, nil
}

func (s *Service) UpdateContact(ctx context.Context, tenantID, customerID, contactID uuid.UUID, in ContactInput) (*ContactDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	c, err := s.contacts.FindByID(ctx, customerID, contactID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.FullName) != "" {
		c.FullName = strings.TrimSpace(in.FullName)
	}
	if strings.TrimSpace(in.Phone) != "" {
		c.Phone = strings.TrimSpace(in.Phone)
	}
	if in.Email != nil {
		c.Email = in.Email
	}
	if in.Position != nil {
		c.Position = in.Position
	}
	c.IsPrimary = in.IsPrimary
	if err := s.contacts.Update(ctx, c); err != nil {
		return nil, err
	}
	dto := toContactDTO(*c)
	return &dto, nil
}

func (s *Service) DeleteContact(ctx context.Context, tenantID, customerID, contactID uuid.UUID) error {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return err
	}
	return s.contacts.SoftDelete(ctx, customerID, contactID)
}

type AddressDTO struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Label      string    `json:"label"`
	Address    string    `json:"address"`
	Lat        *float64  `json:"lat,omitempty"`
	Lng        *float64  `json:"lng,omitempty"`
	IsPrimary  bool      `json:"is_primary"`
}

type AddressInput struct {
	Label     string   `json:"label"`
	Address   string   `json:"address"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	IsPrimary bool     `json:"is_primary"`
}

func toAddressDTO(a domain.CustomerAddress) AddressDTO {
	return AddressDTO{
		ID: a.ID, CustomerID: a.CustomerID, Label: a.Label, Address: a.Address,
		Lat: a.Lat, Lng: a.Lng, IsPrimary: a.IsPrimary,
	}
}

func (s *Service) ListAddresses(ctx context.Context, tenantID, customerID uuid.UUID) ([]AddressDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	rows, err := s.addresses.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}
	out := make([]AddressDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAddressDTO(r))
	}
	return out, nil
}

func (s *Service) CreateAddress(ctx context.Context, tenantID, customerID uuid.UUID, in AddressInput) (*AddressDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	label := strings.TrimSpace(in.Label)
	addr := strings.TrimSpace(in.Address)
	if label == "" || addr == "" {
		return nil, apperrors.ErrValidation
	}
	a := &domain.CustomerAddress{
		CustomerID: customerID, Label: label, Address: addr,
		Lat: in.Lat, Lng: in.Lng, IsPrimary: in.IsPrimary,
	}
	if err := s.addresses.Create(ctx, a); err != nil {
		return nil, err
	}
	dto := toAddressDTO(*a)
	return &dto, nil
}

func (s *Service) UpdateAddress(ctx context.Context, tenantID, customerID, addressID uuid.UUID, in AddressInput) (*AddressDTO, error) {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return nil, err
	}
	a, err := s.addresses.FindByID(ctx, customerID, addressID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Label) != "" {
		a.Label = strings.TrimSpace(in.Label)
	}
	if strings.TrimSpace(in.Address) != "" {
		a.Address = strings.TrimSpace(in.Address)
	}
	if in.Lat != nil {
		a.Lat = in.Lat
	}
	if in.Lng != nil {
		a.Lng = in.Lng
	}
	a.IsPrimary = in.IsPrimary
	if err := s.addresses.Update(ctx, a); err != nil {
		return nil, err
	}
	dto := toAddressDTO(*a)
	return &dto, nil
}

func (s *Service) DeleteAddress(ctx context.Context, tenantID, customerID, addressID uuid.UUID) error {
	if _, err := s.customers.FindByID(ctx, tenantID, customerID); err != nil {
		return err
	}
	return s.addresses.SoftDelete(ctx, customerID, addressID)
}

type CustomerCategoryDTO struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type CustomerCategoryInput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (s *Service) ListCustomerCategories(ctx context.Context, tenantID uuid.UUID) ([]CustomerCategoryDTO, error) {
	rows, err := s.categories.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]CustomerCategoryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, CustomerCategoryDTO{ID: r.ID, Code: r.Code, Name: r.Name})
	}
	return out, nil
}

func (s *Service) CreateCustomerCategory(ctx context.Context, tenantID uuid.UUID, in CustomerCategoryInput) (*CustomerCategoryDTO, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return nil, apperrors.ErrValidation
	}
	c := &domain.CustomerCategory{TenantID: tenantID, Code: code, Name: name}
	if err := s.categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return &CustomerCategoryDTO{ID: c.ID, Code: c.Code, Name: c.Name}, nil
}

func (s *Service) UpdateCustomerCategory(ctx context.Context, tenantID, id uuid.UUID, in CustomerCategoryInput) (*CustomerCategoryDTO, error) {
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
	if err := s.categories.Update(ctx, c); err != nil {
		return nil, err
	}
	return &CustomerCategoryDTO{ID: c.ID, Code: c.Code, Name: c.Name}, nil
}

func (s *Service) DeleteCustomerCategory(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.categories.SoftDelete(ctx, tenantID, id)
}
