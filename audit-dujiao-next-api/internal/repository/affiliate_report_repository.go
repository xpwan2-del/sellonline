package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	AffiliateReportSourceProduct  = "product"
	AffiliateReportSourceDefault  = "default"
	AffiliateReportSourceDisabled = "disabled"
	AffiliateReportSourceMixed    = "mixed"
)

type AffiliateReportRepository interface {
	GetSummary(filter AffiliateReportFilter) (AffiliateReportSummaryRow, error)
	GetCustomerStats(filter AffiliateReportFilter) (AffiliateReportCustomerStatsRow, error)
	GetTrend(filter AffiliateReportFilter) ([]AffiliateReportTrendRow, error)
	GetTopAffiliates(filter AffiliateReportFilter, limit int) ([]AffiliateReportTopAffiliateRow, error)
	GetSourceBreakdown(filter AffiliateReportFilter) ([]AffiliateReportSourceRow, error)
	ListCommissions(filter AffiliateReportCommissionListFilter) ([]models.AffiliateCommission, int64, error)
}

type AffiliateReportFilter struct {
	StartAt            time.Time
	EndAt              time.Time
	AffiliateProfileID uint
	AffiliateKeyword   string
	ProductID          uint
	Status             string
}

type AffiliateReportCommissionListFilter struct {
	AffiliateReportFilter
	Page     int
	PageSize int
}

type AffiliateReportSummaryRow struct {
	TotalSalesAmount         decimal.Decimal
	CommissionBaseAmount     decimal.Decimal
	TotalCommission          decimal.Decimal
	AvailableCommission      decimal.Decimal
	PendingCommission        decimal.Decimal
	WithdrawnCommission      decimal.Decimal
	RejectedCommission       decimal.Decimal
	ValidOrderCount          int64
	ClickCount               int64
	CustomerCount            int64
	NewCustomerOrderCount    int64
	RepeatCustomerOrderCount int64
}

type AffiliateReportCustomerStatsRow struct {
	CustomerCount            int64
	NewCustomerOrderCount    int64
	RepeatCustomerOrderCount int64
}

type AffiliateReportTrendRow struct {
	Date             string
	CommissionAmount decimal.Decimal
}

type AffiliateReportTopAffiliateRow struct {
	AffiliateProfileID uint
	AffiliateCode      string
	Email              string
	DisplayName        string
	CommissionAmount   decimal.Decimal
	ValidOrderCount    int64
	ClickCount         int64
}

type AffiliateReportSourceRow struct {
	Source           string
	SalesAmount      decimal.Decimal
	CommissionAmount decimal.Decimal
}

type GormAffiliateReportRepository struct {
	db *gorm.DB
}

func NewAffiliateReportRepository(db *gorm.DB) *GormAffiliateReportRepository {
	return &GormAffiliateReportRepository{db: db}
}

func (r *GormAffiliateReportRepository) GetSummary(filter AffiliateReportFilter) (AffiliateReportSummaryRow, error) {
	result := AffiliateReportSummaryRow{}
	if r == nil || r.db == nil {
		return result, nil
	}

	sales, err := r.sumAttributedSales(filter)
	if err != nil {
		return result, err
	}
	result.TotalSalesAmount = sales.Round(2)

	commissionQuery := r.commissionAggregateQuery(filter)
	baseExpr, amountExpr := reportCommissionAmountExpr(filter)
	var commissionRow struct {
		Base      decimal.Decimal `gorm:"column:base"`
		Total     decimal.Decimal `gorm:"column:total"`
		Available decimal.Decimal `gorm:"column:available"`
		Pending   decimal.Decimal `gorm:"column:pending"`
		Withdrawn decimal.Decimal `gorm:"column:withdrawn"`
		Rejected  decimal.Decimal `gorm:"column:rejected"`
		Valid     int64           `gorm:"column:valid"`
	}
	if err := commissionQuery.Select(fmt.Sprintf(`
		COALESCE(SUM(%s), 0) AS base,
		COALESCE(SUM(%s), 0) AS total,
		COALESCE(SUM(CASE WHEN affiliate_commissions.status = ? AND affiliate_commissions.withdraw_request_id IS NULL THEN %s ELSE 0 END), 0) AS available,
		COALESCE(SUM(CASE WHEN affiliate_commissions.status = ? THEN %s ELSE 0 END), 0) AS pending,
		COALESCE(SUM(CASE WHEN affiliate_commissions.status = ? THEN %s ELSE 0 END), 0) AS withdrawn,
		COALESCE(SUM(CASE WHEN affiliate_commissions.status = ? THEN %s ELSE 0 END), 0) AS rejected,
		COUNT(DISTINCT CASE WHEN affiliate_commissions.status <> ? THEN affiliate_commissions.order_id END) AS valid
	`, baseExpr, amountExpr, amountExpr, amountExpr, amountExpr, amountExpr),
		constants.AffiliateCommissionStatusAvailable,
		constants.AffiliateCommissionStatusPendingConfirm,
		constants.AffiliateCommissionStatusWithdrawn,
		constants.AffiliateCommissionStatusRejected,
		constants.AffiliateCommissionStatusRejected,
	).Scan(&commissionRow).Error; err != nil {
		return result, err
	}
	result.CommissionBaseAmount = commissionRow.Base.Round(2)
	result.TotalCommission = commissionRow.Total.Round(2)
	result.AvailableCommission = commissionRow.Available.Round(2)
	result.PendingCommission = commissionRow.Pending.Round(2)
	result.WithdrawnCommission = commissionRow.Withdrawn.Round(2)
	result.RejectedCommission = commissionRow.Rejected.Round(2)
	result.ValidOrderCount = commissionRow.Valid

	clickQuery := r.db.Model(&models.AffiliateClick{}).
		Where("affiliate_clicks.created_at >= ? AND affiliate_clicks.created_at < ?", filter.StartAt, filter.EndAt)
	if filter.AffiliateProfileID != 0 {
		clickQuery = clickQuery.Where("affiliate_clicks.affiliate_profile_id = ?", filter.AffiliateProfileID)
	}
	if keyword := strings.TrimSpace(filter.AffiliateKeyword); keyword != "" {
		like := "%" + keyword + "%"
		clickQuery = clickQuery.
			Joins("LEFT JOIN affiliate_profiles ap_click ON ap_click.id = affiliate_clicks.affiliate_profile_id").
			Joins("LEFT JOIN users u_click ON u_click.id = ap_click.user_id").
			Where("(u_click.email LIKE ? OR u_click.display_name LIKE ? OR ap_click.affiliate_code LIKE ?)", like, like, like)
	}
	if err := clickQuery.Count(&result.ClickCount).Error; err != nil {
		return result, err
	}

	return result, nil
}

func (r *GormAffiliateReportRepository) GetCustomerStats(filter AffiliateReportFilter) (AffiliateReportCustomerStatsRow, error) {
	result := AffiliateReportCustomerStatsRow{}
	if r == nil || r.db == nil {
		return result, nil
	}

	customerQuery := r.db.Model(&models.AffiliateCustomerRelation{})
	if filter.AffiliateProfileID != 0 {
		customerQuery = customerQuery.Where("affiliate_customer_relations.affiliate_profile_id = ?", filter.AffiliateProfileID)
	}
	if keyword := strings.TrimSpace(filter.AffiliateKeyword); keyword != "" {
		like := "%" + keyword + "%"
		customerQuery = customerQuery.
			Joins("LEFT JOIN affiliate_profiles ap_customer ON ap_customer.id = affiliate_customer_relations.affiliate_profile_id").
			Joins("LEFT JOIN users u_customer ON u_customer.id = ap_customer.user_id").
			Where("(u_customer.email LIKE ? OR u_customer.display_name LIKE ? OR ap_customer.affiliate_code LIKE ?)", like, like, like)
	}
	if err := customerQuery.Count(&result.CustomerCount).Error; err != nil {
		return result, err
	}

	type orderCustomerRow struct {
		OrderID      uint `gorm:"column:order_id"`
		UserID       uint `gorm:"column:user_id"`
		FirstOrderID uint `gorm:"column:first_order_id"`
	}
	rows := make([]orderCustomerRow, 0)
	query := r.attributedParentOrdersQuery(filter).
		Where("orders.user_id > 0").
		Joins(`LEFT JOIN (
			SELECT user_id, affiliate_profile_id, MIN(id) AS first_order_id
			FROM orders
			WHERE parent_id IS NULL
				AND affiliate_profile_id IS NOT NULL
				AND paid_at IS NOT NULL
				AND user_id > 0
			GROUP BY user_id, affiliate_profile_id
		) first_orders ON first_orders.user_id = orders.user_id AND first_orders.affiliate_profile_id = orders.affiliate_profile_id`).
		Select("orders.id AS order_id, orders.user_id AS user_id, COALESCE(first_orders.first_order_id, 0) AS first_order_id")
	if err := query.Scan(&rows).Error; err != nil {
		return result, err
	}
	for _, row := range rows {
		if row.FirstOrderID != 0 && row.OrderID == row.FirstOrderID {
			result.NewCustomerOrderCount++
			continue
		}
		result.RepeatCustomerOrderCount++
	}
	return result, nil
}

func (r *GormAffiliateReportRepository) GetTrend(filter AffiliateReportFilter) ([]AffiliateReportTrendRow, error) {
	if r == nil || r.db == nil {
		return []AffiliateReportTrendRow{}, nil
	}
	dayExpr := dateGroupExpr(r.db, "orders.paid_at", filter.StartAt.Location(), filter.StartAt)
	_, amountExpr := reportCommissionAmountExpr(filter)
	rows := make([]AffiliateReportTrendRow, 0)
	if err := r.commissionAggregateQuery(filter).
		Select(fmt.Sprintf("%s AS date, COALESCE(SUM(%s), 0) AS commission_amount", dayExpr, amountExpr)).
		Group(dayExpr).
		Order("date ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GormAffiliateReportRepository) GetTopAffiliates(filter AffiliateReportFilter, limit int) ([]AffiliateReportTopAffiliateRow, error) {
	if r == nil || r.db == nil {
		return []AffiliateReportTopAffiliateRow{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	rows := make([]AffiliateReportTopAffiliateRow, 0)
	_, amountExpr := reportCommissionAmountExpr(filter)
	query := r.commissionAggregateQuery(filter).
		Select(fmt.Sprintf(`
			affiliate_profiles.id AS affiliate_profile_id,
			affiliate_profiles.affiliate_code AS affiliate_code,
			users.email AS email,
			users.display_name AS display_name,
			COALESCE(SUM(%s), 0) AS commission_amount,
			COUNT(DISTINCT CASE WHEN affiliate_commissions.status <> ? THEN affiliate_commissions.order_id END) AS valid_order_count,
			COALESCE(clicks.click_count, 0) AS click_count
		`, amountExpr), constants.AffiliateCommissionStatusRejected).
		Joins("JOIN affiliate_profiles ON affiliate_profiles.id = affiliate_commissions.affiliate_profile_id").
		Joins("LEFT JOIN users ON users.id = affiliate_profiles.user_id").
		Joins(`LEFT JOIN (
			SELECT affiliate_profile_id, COUNT(*) AS click_count
			FROM affiliate_clicks
			WHERE created_at >= ? AND created_at < ?
			GROUP BY affiliate_profile_id
		) clicks ON clicks.affiliate_profile_id = affiliate_profiles.id`, filter.StartAt, filter.EndAt).
		Group("affiliate_profiles.id, affiliate_profiles.affiliate_code, users.email, users.display_name, clicks.click_count").
		Order("commission_amount DESC").
		Limit(limit)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GormAffiliateReportRepository) GetSourceBreakdown(filter AffiliateReportFilter) ([]AffiliateReportSourceRow, error) {
	if r == nil || r.db == nil {
		return []AffiliateReportSourceRow{}, nil
	}
	rows := make([]AffiliateReportSourceRow, 0, 3)
	commissionRows := make([]AffiliateReportSourceRow, 0, 2)
	sourceExpr := "CASE WHEN products.affiliate_commission_rate IS NOT NULL THEN '" + AffiliateReportSourceProduct + "' ELSE '" + AffiliateReportSourceDefault + "' END"
	commissionItemQuery := r.commissionScopeQuery(filter, false).
		Joins("JOIN affiliate_commission_items aci ON aci.affiliate_commission_id = affiliate_commissions.id").
		Joins("JOIN products ON products.id = aci.product_id").
		Where("products.is_affiliate_enabled = ?", true)
	if filter.ProductID != 0 {
		commissionItemQuery = commissionItemQuery.Where("aci.product_id = ?", filter.ProductID)
	}
	if err := commissionItemQuery.
		Select(fmt.Sprintf("%s AS source, COALESCE(SUM(aci.base_amount), 0) AS sales_amount, COALESCE(SUM(aci.commission_amount), 0) AS commission_amount", sourceExpr)).
		Group(sourceExpr).
		Scan(&commissionRows).Error; err != nil {
		return nil, err
	}
	rows = append(rows, commissionRows...)

	if strings.TrimSpace(filter.Status) == "" {
		disabledSales, err := r.sumDisabledSales(filter)
		if err != nil {
			return nil, err
		}
		if disabledSales.GreaterThan(decimal.Zero) {
			rows = append(rows, AffiliateReportSourceRow{
				Source:           AffiliateReportSourceDisabled,
				SalesAmount:      disabledSales.Round(2),
				CommissionAmount: decimal.Zero,
			})
		}
	}
	return rows, nil
}

func (r *GormAffiliateReportRepository) ListCommissions(filter AffiliateReportCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	if r == nil || r.db == nil {
		return []models.AffiliateCommission{}, 0, nil
	}

	var total int64
	if err := r.commissionBaseQuery(filter.AffiliateReportFilter).
		Distinct("affiliate_commissions.id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.AffiliateCommission
	query := r.commissionBaseQuery(filter.AffiliateReportFilter).
		Preload("AffiliateProfile").
		Preload("AffiliateProfile.User").
		Preload("Order").
		Preload("Items.OrderItem").
		Preload("Items.Product")
	if filter.ProductID != 0 {
		query = query.Preload("Items", "product_id = ?", filter.ProductID)
	} else {
		query = query.Preload("Items")
	}
	if err := applyPagination(query, filter.Page, filter.PageSize).
		Select("DISTINCT affiliate_commissions.*").
		Order("affiliate_commissions.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *GormAffiliateReportRepository) commissionAggregateQuery(filter AffiliateReportFilter) *gorm.DB {
	if filter.ProductID == 0 {
		return r.commissionScopeQuery(filter, true)
	}
	return r.commissionScopeQuery(filter, false).
		Joins("JOIN affiliate_commission_items acia ON acia.affiliate_commission_id = affiliate_commissions.id").
		Where("acia.product_id = ?", filter.ProductID)
}

func reportCommissionAmountExpr(filter AffiliateReportFilter) (string, string) {
	if filter.ProductID != 0 {
		return "acia.base_amount", "acia.commission_amount"
	}
	return "affiliate_commissions.base_amount", "affiliate_commissions.commission_amount"
}

func (r *GormAffiliateReportRepository) commissionBaseQuery(filter AffiliateReportFilter) *gorm.DB {
	return r.commissionScopeQuery(filter, true)
}

func (r *GormAffiliateReportRepository) commissionScopeQuery(filter AffiliateReportFilter, includeProductFilter bool) *gorm.DB {
	query := r.db.Model(&models.AffiliateCommission{}).
		Joins("JOIN orders ON orders.id = affiliate_commissions.order_id").
		Where("orders.paid_at >= ? AND orders.paid_at < ?", filter.StartAt, filter.EndAt)
	if filter.AffiliateProfileID != 0 {
		query = query.Where("affiliate_commissions.affiliate_profile_id = ?", filter.AffiliateProfileID)
	}
	if keyword := strings.TrimSpace(filter.AffiliateKeyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.
			Joins("LEFT JOIN affiliate_profiles ap_filter ON ap_filter.id = affiliate_commissions.affiliate_profile_id").
			Joins("LEFT JOIN users u_filter ON u_filter.id = ap_filter.user_id").
			Where("(u_filter.email LIKE ? OR u_filter.display_name LIKE ? OR ap_filter.affiliate_code LIKE ?)", like, like, like)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("affiliate_commissions.status = ?", status)
	}
	if includeProductFilter && filter.ProductID != 0 {
		query = query.Joins("JOIN affiliate_commission_items acif ON acif.affiliate_commission_id = affiliate_commissions.id").
			Where("acif.product_id = ?", filter.ProductID)
	}
	return query
}

func (r *GormAffiliateReportRepository) attributedParentOrdersQuery(filter AffiliateReportFilter) *gorm.DB {
	query := r.db.Model(&models.Order{}).
		Where("orders.parent_id IS NULL").
		Where("orders.affiliate_profile_id IS NOT NULL").
		Where("orders.paid_at >= ? AND orders.paid_at < ?", filter.StartAt, filter.EndAt)
	if filter.AffiliateProfileID != 0 {
		query = query.Where("orders.affiliate_profile_id = ?", filter.AffiliateProfileID)
	}
	if keyword := strings.TrimSpace(filter.AffiliateKeyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.
			Joins("LEFT JOIN affiliate_profiles ap_order ON ap_order.id = orders.affiliate_profile_id").
			Joins("LEFT JOIN users u_order ON u_order.id = ap_order.user_id").
			Where("(u_order.email LIKE ? OR u_order.display_name LIKE ? OR ap_order.affiliate_code LIKE ?)", like, like, like)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Joins("JOIN affiliate_commissions acs ON acs.order_id = orders.id").
			Where("acs.status = ?", status)
	}
	return query
}

func (r *GormAffiliateReportRepository) sumAttributedSales(filter AffiliateReportFilter) (decimal.Decimal, error) {
	var total decimal.Decimal
	if filter.ProductID == 0 {
		err := r.attributedParentOrdersQuery(filter).
			Select("COALESCE(SUM(orders.total_amount), 0)").
			Scan(&total).Error
		return total.Round(2), err
	}
	err := r.attributedParentOrdersQuery(filter).
		Joins("JOIN orders child_orders ON child_orders.parent_id = orders.id").
		Joins("JOIN order_items ON order_items.order_id = child_orders.id").
		Where("order_items.product_id = ?", filter.ProductID).
		Select("COALESCE(SUM(order_items.total_price), 0)").
		Scan(&total).Error
	return total.Round(2), err
}

func (r *GormAffiliateReportRepository) sumDisabledSales(filter AffiliateReportFilter) (decimal.Decimal, error) {
	var total decimal.Decimal
	query := r.attributedParentOrdersQuery(filter).
		Joins("JOIN orders child_orders ON child_orders.parent_id = orders.id").
		Joins("JOIN order_items ON order_items.order_id = child_orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.is_affiliate_enabled = ?", false)
	if filter.ProductID != 0 {
		query = query.Where("order_items.product_id = ?", filter.ProductID)
	}
	err := query.Select("COALESCE(SUM(order_items.total_price), 0)").Scan(&total).Error
	return total.Round(2), err
}
