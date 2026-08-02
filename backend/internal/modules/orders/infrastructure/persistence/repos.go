package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepo struct{ db *gorm.DB }

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

func toOrder(m OrderModel) domain.Order {
	return domain.Order{
		ID: m.ID, TenantID: m.TenantID, Number: m.Number, CustomerID: m.CustomerID,
		AgentID: m.AgentID, BranchID: m.BranchID, WarehouseID: m.WarehouseID, VisitID: m.VisitID,
		Status: m.Status, Currency: m.Currency, Subtotal: m.Subtotal, DiscountTotal: m.DiscountTotal,
		TaxTotal: m.TaxTotal, GrandTotal: m.GrandTotal, OrderedAt: m.OrderedAt, DeliveryDate: m.DeliveryDate,
		PriceListID: m.PriceListID, PromotionID: m.PromotionID, Comment: m.Comment,
		ClientRequestID: m.ClientRequestID, Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func toOrderLine(m OrderLineModel) domain.OrderLine {
	return domain.OrderLine{
		ID: m.ID, OrderID: m.OrderID, ProductID: m.ProductID, Qty: m.Qty, UnitPrice: m.UnitPrice,
		Discount: m.Discount, Tax: m.Tax, LineTotal: m.LineTotal, PromotionItemID: m.PromotionItemID,
	}
}

func toHistory(m OrderStatusHistoryModel) domain.OrderStatusHistory {
	return domain.OrderStatusHistory{
		ID: m.ID, OrderID: m.OrderID, FromStatus: m.FromStatus, ToStatus: m.ToStatus,
		ChangedBy: m.ChangedBy, Comment: m.Comment, CreatedAt: m.CreatedAt,
	}
}

func (r *OrderRepo) loadLines(ctx context.Context, tx *gorm.DB, orderID uuid.UUID) ([]domain.OrderLine, error) {
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}
	var rows []OrderLineModel
	if err := db.Where("order_id = ?", orderID).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.OrderLine, 0, len(rows))
	for _, m := range rows {
		out = append(out, toOrderLine(m))
	}
	return out, nil
}

func (r *OrderRepo) List(ctx context.Context, tenantID uuid.UUID, filters domain.OrderListFilters, page, perPage int) ([]domain.Order, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&OrderModel{}).Where("tenant_id = ?", tenantID)
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.CustomerID != nil {
		q = q.Where("customer_id = ?", *filters.CustomerID)
	}
	if filters.AgentID != nil {
		q = q.Where("agent_id = ?", *filters.AgentID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []OrderModel
	if err := q.Order("ordered_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Order, 0, len(rows))
	for _, m := range rows {
		out = append(out, toOrder(m))
	}
	return out, total, nil
}

func (r *OrderRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Order, []domain.OrderLine, error) {
	var m OrderModel
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
	o := toOrder(m)
	return &o, lines, nil
}

func (r *OrderRepo) FindByClientRequestID(ctx context.Context, tenantID uuid.UUID, clientRequestID string) (*domain.Order, []domain.OrderLine, error) {
	var m OrderModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND client_request_id = ?", tenantID, clientRequestID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, apperrors.ErrNotFound
		}
		return nil, nil, err
	}
	lines, err := r.loadLines(ctx, nil, m.ID)
	if err != nil {
		return nil, nil, err
	}
	o := toOrder(m)
	return &o, lines, nil
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order, lines []domain.OrderLine) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now().UTC()
	order.CreatedAt, order.UpdatedAt = now, now
	if order.OrderedAt.IsZero() {
		order.OrderedAt = now
	}
	if order.Version == 0 {
		order.Version = 1
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&OrderModel{
			ID: order.ID, TenantID: order.TenantID, Number: order.Number, CustomerID: order.CustomerID,
			AgentID: order.AgentID, BranchID: order.BranchID, WarehouseID: order.WarehouseID, VisitID: order.VisitID,
			Status: order.Status, Currency: order.Currency, Subtotal: order.Subtotal, DiscountTotal: order.DiscountTotal,
			TaxTotal: order.TaxTotal, GrandTotal: order.GrandTotal, OrderedAt: order.OrderedAt, DeliveryDate: order.DeliveryDate,
			PriceListID: order.PriceListID, PromotionID: order.PromotionID, Comment: order.Comment,
			ClientRequestID: order.ClientRequestID, Version: order.Version, CreatedAt: order.CreatedAt, UpdatedAt: order.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		for i := range lines {
			if lines[i].ID == uuid.Nil {
				lines[i].ID = uuid.New()
			}
			lines[i].OrderID = order.ID
			if err := tx.Create(&OrderLineModel{
				ID: lines[i].ID, OrderID: lines[i].OrderID, ProductID: lines[i].ProductID, Qty: lines[i].Qty,
				UnitPrice: lines[i].UnitPrice, Discount: lines[i].Discount, Tax: lines[i].Tax,
				LineTotal: lines[i].LineTotal, PromotionItemID: lines[i].PromotionItemID,
			}).Error; err != nil {
				return err
			}
		}

		hist := OrderStatusHistoryModel{
			ID: uuid.New(), OrderID: order.ID, FromStatus: "", ToStatus: order.Status, CreatedAt: now,
		}
		return tx.Create(&hist).Error
	})
}

func (r *OrderRepo) Update(ctx context.Context, order *domain.Order, lines []domain.OrderLine) error {
	order.UpdatedAt = time.Now().UTC()
	order.Version++

	updates := map[string]any{
		"number": order.Number, "customer_id": order.CustomerID, "agent_id": order.AgentID,
		"branch_id": order.BranchID, "warehouse_id": order.WarehouseID, "visit_id": order.VisitID,
		"status": order.Status, "currency": order.Currency, "subtotal": order.Subtotal,
		"discount_total": order.DiscountTotal, "tax_total": order.TaxTotal, "grand_total": order.GrandTotal,
		"ordered_at": order.OrderedAt, "delivery_date": order.DeliveryDate, "price_list_id": order.PriceListID,
		"promotion_id": order.PromotionID, "comment": order.Comment, "client_request_id": order.ClientRequestID,
		"version": order.Version, "updated_at": order.UpdatedAt,
	}

	if lines == nil {
		return r.db.WithContext(ctx).Model(&OrderModel{}).
			Where("id = ? AND tenant_id = ?", order.ID, order.TenantID).
			Updates(updates).Error
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&OrderModel{}).Where("id = ? AND tenant_id = ?", order.ID, order.TenantID).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apperrors.ErrNotFound
		}
		if err := tx.Where("order_id = ?", order.ID).Delete(&OrderLineModel{}).Error; err != nil {
			return err
		}
		for i := range lines {
			if lines[i].ID == uuid.Nil {
				lines[i].ID = uuid.New()
			}
			lines[i].OrderID = order.ID
			if err := tx.Create(&OrderLineModel{
				ID: lines[i].ID, OrderID: lines[i].OrderID, ProductID: lines[i].ProductID, Qty: lines[i].Qty,
				UnitPrice: lines[i].UnitPrice, Discount: lines[i].Discount, Tax: lines[i].Tax,
				LineTotal: lines[i].LineTotal, PromotionItemID: lines[i].PromotionItemID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepo) AddHistory(ctx context.Context, h *domain.OrderStatusHistory) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(&OrderStatusHistoryModel{
		ID: h.ID, OrderID: h.OrderID, FromStatus: h.FromStatus, ToStatus: h.ToStatus,
		ChangedBy: h.ChangedBy, Comment: h.Comment, CreatedAt: h.CreatedAt,
	}).Error
}

func (r *OrderRepo) ListHistory(ctx context.Context, orderID uuid.UUID) ([]domain.OrderStatusHistory, error) {
	var rows []OrderStatusHistoryModel
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.OrderStatusHistory, 0, len(rows))
	for _, m := range rows {
		out = append(out, toHistory(m))
	}
	return out, nil
}

func (r *OrderRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&OrderModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
