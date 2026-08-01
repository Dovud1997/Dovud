package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/returns/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReturnRepo struct{ db *gorm.DB }

func NewReturnRepo(db *gorm.DB) *ReturnRepo { return &ReturnRepo{db: db} }

func toReturn(m ReturnModel) domain.Return {
	return domain.Return{
		ID: m.ID, TenantID: m.TenantID, Number: m.Number, OrderID: m.OrderID,
		CustomerID: m.CustomerID, AgentID: m.AgentID, Status: m.Status, Reason: m.Reason,
		Currency: m.Currency, Subtotal: m.Subtotal, TaxTotal: m.TaxTotal, GrandTotal: m.GrandTotal,
		Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func toReturnLine(m ReturnLineModel) domain.ReturnLine {
	return domain.ReturnLine{
		ID: m.ID, ReturnID: m.ReturnID, ProductID: m.ProductID, Qty: m.Qty,
		UnitPrice: m.UnitPrice, LineTotal: m.LineTotal, Reason: m.Reason,
	}
}

func (r *ReturnRepo) loadLines(ctx context.Context, tx *gorm.DB, returnID uuid.UUID) ([]domain.ReturnLine, error) {
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}
	var rows []ReturnLineModel
	if err := db.Where("return_id = ?", returnID).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ReturnLine, 0, len(rows))
	for _, m := range rows {
		out = append(out, toReturnLine(m))
	}
	return out, nil
}

func (r *ReturnRepo) List(ctx context.Context, tenantID uuid.UUID, filters domain.ReturnListFilters, page, perPage int) ([]domain.Return, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&ReturnModel{}).Where("tenant_id = ?", tenantID)
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ReturnModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Return, 0, len(rows))
	for _, m := range rows {
		out = append(out, toReturn(m))
	}
	return out, total, nil
}

func (r *ReturnRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Return, []domain.ReturnLine, error) {
	var m ReturnModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, apperrors.ErrNotFound
		}
		return nil, nil, err
	}
	lines, err := r.loadLines(ctx, nil, m.ID)
	if err != nil {
		return nil, nil, err
	}
	ret := toReturn(m)
	return &ret, lines, nil
}

func (r *ReturnRepo) Create(ctx context.Context, ret *domain.Return, lines []domain.ReturnLine) error {
	if ret.ID == uuid.Nil {
		ret.ID = uuid.New()
	}
	now := time.Now().UTC()
	ret.CreatedAt, ret.UpdatedAt = now, now
	if ret.Version == 0 {
		ret.Version = 1
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ReturnModel{
			ID: ret.ID, TenantID: ret.TenantID, Number: ret.Number, OrderID: ret.OrderID,
			CustomerID: ret.CustomerID, AgentID: ret.AgentID, Status: ret.Status, Reason: ret.Reason,
			Currency: ret.Currency, Subtotal: ret.Subtotal, TaxTotal: ret.TaxTotal, GrandTotal: ret.GrandTotal,
			Version: ret.Version, CreatedAt: ret.CreatedAt, UpdatedAt: ret.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		for i := range lines {
			if lines[i].ID == uuid.Nil {
				lines[i].ID = uuid.New()
			}
			lines[i].ReturnID = ret.ID
			if err := tx.Create(&ReturnLineModel{
				ID: lines[i].ID, ReturnID: lines[i].ReturnID, ProductID: lines[i].ProductID,
				Qty: lines[i].Qty, UnitPrice: lines[i].UnitPrice, LineTotal: lines[i].LineTotal,
				Reason: lines[i].Reason,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ReturnRepo) Update(ctx context.Context, ret *domain.Return, lines []domain.ReturnLine) error {
	ret.UpdatedAt = time.Now().UTC()
	ret.Version++

	updates := map[string]any{
		"number": ret.Number, "order_id": ret.OrderID, "customer_id": ret.CustomerID,
		"agent_id": ret.AgentID, "status": ret.Status, "reason": ret.Reason,
		"currency": ret.Currency, "subtotal": ret.Subtotal, "tax_total": ret.TaxTotal,
		"grand_total": ret.GrandTotal, "version": ret.Version, "updated_at": ret.UpdatedAt,
	}

	if lines == nil {
		return r.db.WithContext(ctx).Model(&ReturnModel{}).
			Where("id = ? AND tenant_id = ?", ret.ID, ret.TenantID).
			Updates(updates).Error
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&ReturnModel{}).Where("id = ? AND tenant_id = ?", ret.ID, ret.TenantID).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apperrors.ErrNotFound
		}
		if err := tx.Where("return_id = ?", ret.ID).Delete(&ReturnLineModel{}).Error; err != nil {
			return err
		}
		for i := range lines {
			if lines[i].ID == uuid.Nil {
				lines[i].ID = uuid.New()
			}
			lines[i].ReturnID = ret.ID
			if err := tx.Create(&ReturnLineModel{
				ID: lines[i].ID, ReturnID: lines[i].ReturnID, ProductID: lines[i].ProductID,
				Qty: lines[i].Qty, UnitPrice: lines[i].UnitPrice, LineTotal: lines[i].LineTotal,
				Reason: lines[i].Reason,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ReturnRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&ReturnModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
