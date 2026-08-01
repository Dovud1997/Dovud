package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReceivableRepo struct{ db *gorm.DB }

func NewReceivableRepo(db *gorm.DB) *ReceivableRepo { return &ReceivableRepo{db: db} }

func toReceivable(m ReceivableModel) domain.Receivable {
	return domain.Receivable{
		ID: m.ID, TenantID: m.TenantID, CustomerID: m.CustomerID, DocumentType: m.DocumentType,
		DocumentID: m.DocumentID, Amount: m.Amount, PaidAmount: m.PaidAmount, Balance: m.Balance,
		DueDate: m.DueDate, Status: m.Status, Currency: m.Currency, Version: m.Version,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func toPayment(m ReceivablePaymentModel) domain.ReceivablePayment {
	return domain.ReceivablePayment{
		ID: m.ID, ReceivableID: m.ReceivableID, Amount: m.Amount, PaidAt: m.PaidAt,
		Method: m.Method, Reference: m.Reference, CreatedBy: m.CreatedBy, CreatedAt: m.CreatedAt,
	}
}

func (r *ReceivableRepo) List(ctx context.Context, tenantID uuid.UUID, filters domain.ReceivableListFilters, page, perPage int) ([]domain.Receivable, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&ReceivableModel{}).Where("tenant_id = ?", tenantID)
	if filters.CustomerID != nil {
		q = q.Where("customer_id = ?", *filters.CustomerID)
	}
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ReceivableModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Receivable, 0, len(rows))
	for _, m := range rows {
		out = append(out, toReceivable(m))
	}
	return out, total, nil
}

func (r *ReceivableRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Receivable, error) {
	var m ReceivableModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	rec := toReceivable(m)
	return &rec, nil
}

func (r *ReceivableRepo) Create(ctx context.Context, rec *domain.Receivable) error {
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	now := time.Now().UTC()
	rec.CreatedAt, rec.UpdatedAt = now, now
	if rec.Version == 0 {
		rec.Version = 1
	}
	return r.db.WithContext(ctx).Create(&ReceivableModel{
		ID: rec.ID, TenantID: rec.TenantID, CustomerID: rec.CustomerID, DocumentType: rec.DocumentType,
		DocumentID: rec.DocumentID, Amount: rec.Amount, PaidAmount: rec.PaidAmount, Balance: rec.Balance,
		DueDate: rec.DueDate, Status: rec.Status, Currency: rec.Currency, Version: rec.Version,
		CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}).Error
}

func (r *ReceivableRepo) Update(ctx context.Context, rec *domain.Receivable) error {
	rec.UpdatedAt = time.Now().UTC()
	rec.Version++
	return r.db.WithContext(ctx).Model(&ReceivableModel{}).Where("id = ? AND tenant_id = ?", rec.ID, rec.TenantID).Updates(map[string]any{
		"customer_id": rec.CustomerID, "document_type": rec.DocumentType, "document_id": rec.DocumentID,
		"amount": rec.Amount, "paid_amount": rec.PaidAmount, "balance": rec.Balance,
		"due_date": rec.DueDate, "status": rec.Status, "currency": rec.Currency,
		"version": rec.Version, "updated_at": rec.UpdatedAt,
	}).Error
}

func (r *ReceivableRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&ReceivableModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ReceivableRepo) ListPayments(ctx context.Context, receivableID uuid.UUID) ([]domain.ReceivablePayment, error) {
	var rows []ReceivablePaymentModel
	if err := r.db.WithContext(ctx).Where("receivable_id = ?", receivableID).Order("paid_at ASC, created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ReceivablePayment, 0, len(rows))
	for _, m := range rows {
		out = append(out, toPayment(m))
	}
	return out, nil
}

func (r *ReceivableRepo) AddPayment(ctx context.Context, p *domain.ReceivablePayment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.PaidAt.IsZero() {
		p.PaidAt = p.CreatedAt
	}
	return r.db.WithContext(ctx).Create(&ReceivablePaymentModel{
		ID: p.ID, ReceivableID: p.ReceivableID, Amount: p.Amount, PaidAt: p.PaidAt,
		Method: p.Method, Reference: p.Reference, CreatedBy: p.CreatedBy, CreatedAt: p.CreatedAt,
	}).Error
}

func (r *ReceivableRepo) AgingReport(ctx context.Context, tenantID uuid.UUID) (*domain.AgingReport, error) {
	var rows []ReceivableModel
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status != ? AND balance > 0", tenantID, domain.StatusClosed).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	report := &domain.AgingReport{}
	for _, m := range rows {
		daysPastDue := 0
		if m.DueDate != nil {
			due := time.Date(m.DueDate.Year(), m.DueDate.Month(), m.DueDate.Day(), 0, 0, 0, 0, time.UTC)
			if today.After(due) {
				daysPastDue = int(today.Sub(due).Hours() / 24)
			}
		}
		switch {
		case daysPastDue <= 30:
			report.Bucket0To30 += m.Balance
		case daysPastDue <= 60:
			report.Bucket31To60 += m.Balance
		case daysPastDue <= 90:
			report.Bucket61To90 += m.Balance
		default:
			report.Bucket90Plus += m.Balance
		}
	}
	return report, nil
}

func (r *ReceivableRepo) SumBalanceByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&ReceivableModel{}).
		Where("tenant_id = ? AND customer_id = ? AND status != ?", tenantID, customerID, domain.StatusClosed).
		Select("COALESCE(SUM(balance), 0)").Scan(&sum).Error
	return sum, err
}

type CreditLimitRepo struct{ db *gorm.DB }

func NewCreditLimitRepo(db *gorm.DB) *CreditLimitRepo { return &CreditLimitRepo{db: db} }

func toCreditLimit(m CreditLimitModel) domain.CreditLimit {
	return domain.CreditLimit{
		ID: m.ID, TenantID: m.TenantID, CustomerID: m.CustomerID,
		Amount: m.Amount, Currency: m.Currency, UpdatedAt: m.UpdatedAt,
	}
}

func (r *CreditLimitRepo) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.CreditLimit, error) {
	var m CreditLimitModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	cl := toCreditLimit(m)
	return &cl, nil
}

func (r *CreditLimitRepo) Upsert(ctx context.Context, cl *domain.CreditLimit) error {
	cl.UpdatedAt = time.Now().UTC()
	var existing CreditLimitModel
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND customer_id = ?", cl.TenantID, cl.CustomerID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if cl.ID == uuid.Nil {
			cl.ID = uuid.New()
		}
		return r.db.WithContext(ctx).Create(&CreditLimitModel{
			ID: cl.ID, TenantID: cl.TenantID, CustomerID: cl.CustomerID,
			Amount: cl.Amount, Currency: cl.Currency, UpdatedAt: cl.UpdatedAt,
		}).Error
	}
	if err != nil {
		return err
	}
	cl.ID = existing.ID
	return r.db.WithContext(ctx).Model(&CreditLimitModel{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"amount": cl.Amount, "currency": cl.Currency, "updated_at": cl.UpdatedAt,
	}).Error
}
