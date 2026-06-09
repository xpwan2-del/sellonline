package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"gorm.io/gorm"
)

// AffiliateCustomerRelationListInput 代理商客户列表输入
type AffiliateCustomerRelationListInput struct {
	Page               int
	PageSize           int
	AffiliateProfileID uint
	Keyword            string
	RegisteredFrom     *time.Time
	RegisteredTo       *time.Time
}

// AffiliateCustomerSummary 当前一级代理商名下客户展示数据。
type AffiliateCustomerSummary struct {
	ID               uint
	Email            string
	DisplayName      string
	RegisteredAt     time.Time
	LastOrderAt      *time.Time
	OrderCount       int64
	CommissionAmount models.Money
}

// ListCustomerRelations 查询一级代理商名下买家关系
func (s *AffiliateService) ListCustomerRelations(input AffiliateCustomerRelationListInput) ([]models.AffiliateCustomerRelation, int64, error) {
	if s.repo == nil {
		return nil, 0, ErrNotFound
	}
	return s.repo.ListCustomerRelations(repository.AffiliateCustomerRelationListFilter{
		Page:               input.Page,
		PageSize:           input.PageSize,
		AffiliateProfileID: input.AffiliateProfileID,
		Keyword:            strings.TrimSpace(input.Keyword),
		RegisteredFrom:     input.RegisteredFrom,
		RegisteredTo:       input.RegisteredTo,
	})
}

// ListMyCustomers 查询当前登录一级代理商名下客户。
func (s *AffiliateService) ListMyCustomers(userID uint, input AffiliateCustomerRelationListInput) ([]AffiliateCustomerSummary, int64, error) {
	if userID == 0 || s.repo == nil {
		return []AffiliateCustomerSummary{}, 0, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return []AffiliateCustomerSummary{}, 0, nil
	}
	rows, total, err := s.repo.ListCustomerSummaries(repository.AffiliateCustomerRelationListFilter{
		Page:               input.Page,
		PageSize:           input.PageSize,
		AffiliateProfileID: profile.ID,
		Keyword:            strings.TrimSpace(input.Keyword),
		RegisteredFrom:     input.RegisteredFrom,
		RegisteredTo:       input.RegisteredTo,
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]AffiliateCustomerSummary, 0, len(rows))
	for i := range rows {
		result = append(result, AffiliateCustomerSummary{
			ID:               rows[i].ID,
			Email:            rows[i].Email,
			DisplayName:      rows[i].DisplayName,
			RegisteredAt:     rows[i].RegisteredAt,
			LastOrderAt:      rows[i].LastOrderAt,
			OrderCount:       rows[i].OrderCount,
			CommissionAmount: models.NewMoneyFromDecimal(rows[i].CommissionAmount),
		})
	}
	return result, total, nil
}

// BindCustomerRelationForOrderTx 在订单创建成功的事务中绑定买家长期归属。
func (s *AffiliateService) BindCustomerRelationForOrderTx(tx *gorm.DB, userID uint, profileID uint, affiliateCode string, orderID uint, now time.Time) error {
	if s == nil || s.repo == nil || tx == nil || userID == 0 || profileID == 0 {
		return nil
	}
	repo := s.repo.WithTx(tx)
	existing, err := repo.GetCustomerRelationByUserID(userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	orderIDPtr := orderID
	relation := &models.AffiliateCustomerRelation{
		CustomerUserID:      userID,
		AffiliateProfileID:  profileID,
		SourceAffiliateCode: normalizeAffiliateCode(affiliateCode),
		SourceOrderID:       &orderIDPtr,
		SourceType:          constants.AffiliateCustomerRelationSourceOrder,
		BoundAt:             now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return repo.CreateCustomerRelation(relation)
}
