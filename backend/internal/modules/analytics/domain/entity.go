package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"

	ScopeTenant = "tenant"
	ScopeBranch = "branch"
	ScopeAgent  = "agent"
)

type KpiDefinition struct {
	ID          uuid.UUID
	TenantID    *uuid.UUID
	Code        string
	Name        string
	Description string
	Unit        string
}

type KpiSnapshot struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	KpiCode     string
	Period      string
	PeriodStart time.Time
	ScopeType   string
	ScopeID     *uuid.UUID
	Value       float64
	CreatedAt   time.Time
}

type DashboardSummary struct {
	OrdersToday            int64   `json:"orders_today"`
	OrdersTotalAmountToday float64 `json:"orders_total_amount_today"`
	VisitsToday            int64   `json:"visits_today"`
	OpenReceivables        float64 `json:"open_receivables"`
	ActiveAgents           int64   `json:"active_agents"`
	PendingOrders          int64   `json:"pending_orders"`
}

func ValidPeriod(p string) bool {
	switch p {
	case PeriodDaily, PeriodWeekly, PeriodMonthly:
		return true
	default:
		return false
	}
}

func ValidScope(s string) bool {
	switch s {
	case ScopeTenant, ScopeBranch, ScopeAgent:
		return true
	default:
		return false
	}
}
