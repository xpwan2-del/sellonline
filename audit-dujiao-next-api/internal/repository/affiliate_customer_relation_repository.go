package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AffiliateCustomerSummaryRow 当前一级代理商名下客户展示行。
type AffiliateCustomerSummaryRow struct {
	ID               uint            `gorm:"column:id"`
	Email            string          `gorm:"column:email"`
	DisplayName      string          `gorm:"column:display_name"`
	RegisteredAt     time.Time       `gorm:"column:registered_at"`
	LastOrderAt      *time.Time      `gorm:"-"`
	OrderCount       int64           `gorm:"column:order_count"`
	CommissionAmount decimal.Decimal `gorm:"column:commission_amount"`
}

type affiliateCustomerSummaryScanRow struct {
	ID               uint            `gorm:"column:id"`
	Email            string          `gorm:"column:email"`
	DisplayName      string          `gorm:"column:display_name"`
	RegisteredAt     time.Time       `gorm:"column:registered_at"`
	LastOrderAt      string          `gorm:"column:last_order_at"`
	OrderCount       int64           `gorm:"column:order_count"`
	CommissionAmount decimal.Decimal `gorm:"column:commission_amount"`
}

// GetCustomerRelationByUserID 查询买家的长期代理归属
func (r *GormAffiliateRepository) GetCustomerRelationByUserID(userID uint) (*models.AffiliateCustomerRelation, error) {
	if userID == 0 {
		return nil, nil
	}
	var row models.AffiliateCustomerRelation
	if err := r.db.Preload("AffiliateProfile.User").
		Where("customer_user_id = ?", userID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// CreateCustomerRelation 创建买家的长期代理归属；已有归属时不覆盖
func (r *GormAffiliateRepository) CreateCustomerRelation(relation *models.AffiliateCustomerRelation) error {
	if relation == nil || relation.CustomerUserID == 0 || relation.AffiliateProfileID == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "customer_user_id"}},
		DoNothing: true,
	}).Create(relation).Error
}

// ListCustomerRelations 查询一级代理商名下买家关系
func (r *GormAffiliateRepository) ListCustomerRelations(filter AffiliateCustomerRelationListFilter) ([]models.AffiliateCustomerRelation, int64, error) {
	query := r.db.Model(&models.AffiliateCustomerRelation{}).
		Preload("Customer").
		Preload("AffiliateProfile.User")
	if filter.AffiliateProfileID != 0 {
		query = query.Where("affiliate_profile_id = ?", filter.AffiliateProfileID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Joins("LEFT JOIN users ON users.id = affiliate_customer_relations.customer_user_id").
			Where("(users.email LIKE ? OR users.display_name LIKE ?)", like, like)
	}
	if filter.RegisteredFrom != nil || filter.RegisteredTo != nil {
		query = query.Joins("LEFT JOIN users registered_users ON registered_users.id = affiliate_customer_relations.customer_user_id")
		if filter.RegisteredFrom != nil {
			query = query.Where("registered_users.created_at >= ?", *filter.RegisteredFrom)
		}
		if filter.RegisteredTo != nil {
			query = query.Where("registered_users.created_at < ?", *filter.RegisteredTo)
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = applyPagination(query, filter.Page, filter.PageSize)
	var rows []models.AffiliateCustomerRelation
	if err := query.Order("affiliate_customer_relations.id desc").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListCustomerSummaries 查询当前一级代理商名下客户及其返佣贡献。
func (r *GormAffiliateRepository) ListCustomerSummaries(filter AffiliateCustomerRelationListFilter) ([]AffiliateCustomerSummaryRow, int64, error) {
	if r == nil || r.db == nil || filter.AffiliateProfileID == 0 {
		return []AffiliateCustomerSummaryRow{}, 0, nil
	}
	query := r.db.Model(&models.AffiliateCustomerRelation{}).
		Joins("JOIN users ON users.id = affiliate_customer_relations.customer_user_id").
		Where("affiliate_customer_relations.affiliate_profile_id = ?", filter.AffiliateProfileID).
		Where("users.deleted_at IS NULL")
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(users.email LIKE ? OR users.display_name LIKE ?)", like, like)
	}
	if filter.RegisteredFrom != nil {
		query = query.Where("users.created_at >= ?", *filter.RegisteredFrom)
	}
	if filter.RegisteredTo != nil {
		query = query.Where("users.created_at < ?", *filter.RegisteredTo)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectExpr := `
		users.id AS id,
		users.email AS email,
		users.display_name AS display_name,
		users.created_at AS registered_at,
		(
			SELECT MAX(orders.paid_at)
			FROM affiliate_commissions
			JOIN orders ON orders.id = affiliate_commissions.order_id
			WHERE affiliate_commissions.affiliate_profile_id = ?
				AND affiliate_commissions.status <> ?
				AND orders.user_id = users.id
		) AS last_order_at,
		(
			SELECT COUNT(DISTINCT affiliate_commissions.order_id)
			FROM affiliate_commissions
			JOIN orders ON orders.id = affiliate_commissions.order_id
			WHERE affiliate_commissions.affiliate_profile_id = ?
				AND affiliate_commissions.status <> ?
				AND orders.user_id = users.id
		) AS order_count,
		(
			SELECT COALESCE(SUM(affiliate_commissions.commission_amount), 0)
			FROM affiliate_commissions
			JOIN orders ON orders.id = affiliate_commissions.order_id
			WHERE affiliate_commissions.affiliate_profile_id = ?
				AND affiliate_commissions.status <> ?
				AND orders.user_id = users.id
		) AS commission_amount
	`

	scanRows := make([]affiliateCustomerSummaryScanRow, 0)
	if err := applyPagination(query, filter.Page, filter.PageSize).
		Select(selectExpr,
			filter.AffiliateProfileID, constants.AffiliateCommissionStatusRejected,
			filter.AffiliateProfileID, constants.AffiliateCommissionStatusRejected,
			filter.AffiliateProfileID, constants.AffiliateCommissionStatusRejected,
		).
		Order("affiliate_customer_relations.id desc").
		Scan(&scanRows).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]AffiliateCustomerSummaryRow, 0, len(scanRows))
	for i := range scanRows {
		rows = append(rows, AffiliateCustomerSummaryRow{
			ID:               scanRows[i].ID,
			Email:            scanRows[i].Email,
			DisplayName:      scanRows[i].DisplayName,
			RegisteredAt:     scanRows[i].RegisteredAt,
			LastOrderAt:      parseAffiliateCustomerSummaryTime(scanRows[i].LastOrderAt),
			OrderCount:       scanRows[i].OrderCount,
			CommissionAmount: scanRows[i].CommissionAmount,
		})
	}
	return rows, total, nil
}

func parseAffiliateCustomerSummaryTime(raw string) *time.Time {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			return &parsed
		}
	}
	return nil
}
