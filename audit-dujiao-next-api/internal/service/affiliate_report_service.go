package service

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/shopspring/decimal"
)

const affiliateReportDefaultRangeDays = 7

const (
	AffiliateReportSourceProduct  = repository.AffiliateReportSourceProduct
	AffiliateReportSourceDefault  = repository.AffiliateReportSourceDefault
	AffiliateReportSourceDisabled = repository.AffiliateReportSourceDisabled
	AffiliateReportSourceMixed    = repository.AffiliateReportSourceMixed
)

// AffiliateReportService 提供推广返利报表查询。
type AffiliateReportService struct {
	reportRepo    repository.AffiliateReportRepository
	affiliateRepo repository.AffiliateRepository
	userRepo      repository.UserRepository
}

func NewAffiliateReportService(
	reportRepo repository.AffiliateReportRepository,
	affiliateRepo repository.AffiliateRepository,
	userRepo repository.UserRepository,
) *AffiliateReportService {
	return &AffiliateReportService{
		reportRepo:    reportRepo,
		affiliateRepo: affiliateRepo,
		userRepo:      userRepo,
	}
}

type AffiliateReportQuery struct {
	StartAt            *time.Time
	EndAt              *time.Time
	AffiliateProfileID uint
	AffiliateKeyword   string
	ProductID          uint
	Status             string
}

type AffiliateReportCommissionQuery struct {
	AffiliateReportQuery
	Page     int
	PageSize int
}

type AffiliateReportResponse struct {
	Summary         AffiliateReportSummary           `json:"summary"`
	Trend           []AffiliateReportTrendPoint      `json:"trend"`
	TopAffiliates   []AffiliateReportTopAffiliate    `json:"top_affiliates,omitempty"`
	SourceBreakdown []AffiliateReportSourceBreakdown `json:"source_breakdown"`
}

type AffiliateReportSummary struct {
	AffiliateCode            string `json:"affiliate_code,omitempty"`
	PromotionURL             string `json:"promotion_url,omitempty"`
	TotalSalesAmount         string `json:"total_sales_amount,omitempty"`
	CommissionBaseAmount     string `json:"commission_base_amount"`
	TotalCommission          string `json:"total_commission"`
	AvailableCommission      string `json:"available_commission"`
	PendingCommission        string `json:"pending_commission"`
	WithdrawnCommission      string `json:"withdrawn_commission"`
	RejectedCommission       string `json:"rejected_commission"`
	ValidOrderCount          int64  `json:"valid_order_count"`
	ClickCount               int64  `json:"click_count"`
	CustomerCount            int64  `json:"customer_count"`
	NewCustomerOrderCount    int64  `json:"new_customer_order_count"`
	RepeatCustomerOrderCount int64  `json:"repeat_customer_order_count"`
	ConversionRate           string `json:"conversion_rate"`
	AverageCommission        string `json:"average_commission"`
	AverageCommissionRate    string `json:"average_commission_rate"`
}

type AffiliateReportTrendPoint struct {
	Date             string `json:"date"`
	CommissionAmount string `json:"commission_amount"`
}

type AffiliateReportTopAffiliate struct {
	AffiliateProfileID uint   `json:"affiliate_profile_id"`
	AffiliateCode      string `json:"affiliate_code"`
	Email              string `json:"email"`
	DisplayName        string `json:"display_name"`
	CommissionAmount   string `json:"commission_amount"`
	ValidOrderCount    int64  `json:"valid_order_count"`
	ClickCount         int64  `json:"click_count"`
	ConversionRate     string `json:"conversion_rate"`
}

type AffiliateReportSourceBreakdown struct {
	Source           string `json:"source"`
	Label            string `json:"label"`
	SalesAmount      string `json:"sales_amount"`
	CommissionAmount string `json:"commission_amount"`
}

type AffiliateReportCommission struct {
	ID                 uint                            `json:"id"`
	AffiliateProfileID uint                            `json:"affiliate_profile_id,omitempty"`
	AffiliateCode      string                          `json:"affiliate_code,omitempty"`
	PromoterEmail      string                          `json:"promoter_email,omitempty"`
	PromoterName       string                          `json:"promoter_name,omitempty"`
	OrderID            uint                            `json:"order_id"`
	OrderNo            string                          `json:"order_no"`
	BuyerEmail         string                          `json:"buyer_email"`
	BuyerMasked        string                          `json:"buyer_masked"`
	OrderTotalAmount   models.Money                    `json:"order_total_amount"`
	BaseAmount         models.Money                    `json:"base_amount"`
	RatePercent        models.Money                    `json:"rate_percent"`
	CommissionAmount   models.Money                    `json:"commission_amount"`
	SourceType         string                          `json:"source_type"`
	SourceLabel        string                          `json:"source_label"`
	Status             string                          `json:"status"`
	CreatedAt          time.Time                       `json:"created_at"`
	AvailableAt        *time.Time                      `json:"available_at,omitempty"`
	CommissionItems    []AffiliateReportCommissionItem `json:"commission_items"`
}

type AffiliateReportCommissionItem struct {
	ID               uint         `json:"id,omitempty"`
	OrderItemID      uint         `json:"order_item_id,omitempty"`
	ProductID        uint         `json:"product_id"`
	ProductTitle     models.JSON  `json:"product_title,omitempty"`
	SKUSnapshot      models.JSON  `json:"sku_snapshot,omitempty"`
	Quantity         int          `json:"quantity"`
	BaseAmount       models.Money `json:"base_amount"`
	RatePercent      models.Money `json:"rate_percent"`
	CommissionAmount models.Money `json:"commission_amount"`
	SourceType       string       `json:"source_type"`
	SourceLabel      string       `json:"source_label"`
}

func (s *AffiliateReportService) GetAdminSummary(query AffiliateReportQuery) (AffiliateReportResponse, error) {
	filter, err := normalizeAffiliateReportFilter(query)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	return s.buildReport(filter, true, "")
}

func (s *AffiliateReportService) GetUserSummary(userID uint, query AffiliateReportQuery, origin string) (AffiliateReportResponse, error) {
	filter, code, opened, err := s.userFilter(userID, query)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	if !opened {
		return emptyAffiliateReportResponse(), nil
	}
	resp, err := s.buildReport(filter, false, code)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	resp.Summary.AffiliateCode = code
	if code != "" {
		resp.Summary.PromotionURL = strings.TrimRight(strings.TrimSpace(origin), "/") + "/?aff=" + code
		if strings.TrimSpace(origin) == "" {
			resp.Summary.PromotionURL = "/?aff=" + code
		}
	}
	return resp, nil
}

func (s *AffiliateReportService) ListAdminCommissions(query AffiliateReportCommissionQuery) ([]AffiliateReportCommission, int64, error) {
	filter, err := normalizeAffiliateReportFilter(query.AffiliateReportQuery)
	if err != nil {
		return nil, 0, err
	}
	return s.listCommissions(repository.AffiliateReportCommissionListFilter{
		AffiliateReportFilter: filter,
		Page:                  query.Page,
		PageSize:              query.PageSize,
	}, true)
}

func (s *AffiliateReportService) ListUserCommissions(userID uint, query AffiliateReportCommissionQuery) ([]AffiliateReportCommission, int64, error) {
	filter, _, opened, err := s.userFilter(userID, query.AffiliateReportQuery)
	if err != nil {
		return nil, 0, err
	}
	if !opened {
		return []AffiliateReportCommission{}, 0, nil
	}
	return s.listCommissions(repository.AffiliateReportCommissionListFilter{
		AffiliateReportFilter: filter,
		Page:                  query.Page,
		PageSize:              query.PageSize,
	}, false)
}

func (s *AffiliateReportService) userFilter(userID uint, query AffiliateReportQuery) (repository.AffiliateReportFilter, string, bool, error) {
	filter, err := normalizeAffiliateReportFilter(query)
	if err != nil {
		return filter, "", false, err
	}
	if userID == 0 || s.affiliateRepo == nil {
		return filter, "", false, nil
	}
	profile, err := s.affiliateRepo.GetProfileByUserID(userID)
	if err != nil {
		return filter, "", false, err
	}
	if profile == nil {
		return filter, "", false, nil
	}
	filter.AffiliateProfileID = profile.ID
	return filter, profile.AffiliateCode, true, nil
}

func (s *AffiliateReportService) buildReport(filter repository.AffiliateReportFilter, includeTop bool, code string) (AffiliateReportResponse, error) {
	if s == nil || s.reportRepo == nil {
		return AffiliateReportResponse{
			Summary: AffiliateReportSummary{
				AffiliateCode: code,
			},
		}, nil
	}
	summaryRow, err := s.reportRepo.GetSummary(filter)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	customerStats, err := s.reportRepo.GetCustomerStats(filter)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	summaryRow.CustomerCount = customerStats.CustomerCount
	summaryRow.NewCustomerOrderCount = customerStats.NewCustomerOrderCount
	summaryRow.RepeatCustomerOrderCount = customerStats.RepeatCustomerOrderCount
	trendRows, err := s.reportRepo.GetTrend(filter)
	if err != nil {
		return AffiliateReportResponse{}, err
	}
	sourceRows, err := s.reportRepo.GetSourceBreakdown(filter)
	if err != nil {
		return AffiliateReportResponse{}, err
	}

	resp := AffiliateReportResponse{
		Summary:         buildAffiliateReportSummary(summaryRow),
		Trend:           buildAffiliateReportTrend(trendRows, filter.StartAt, filter.EndAt),
		SourceBreakdown: buildAffiliateReportSources(sourceRows),
	}
	resp.Summary.AffiliateCode = code
	if includeTop {
		topRows, err := s.reportRepo.GetTopAffiliates(filter, 10)
		if err != nil {
			return AffiliateReportResponse{}, err
		}
		resp.TopAffiliates = buildAffiliateReportTopAffiliates(topRows)
	}
	return resp, nil
}

func (s *AffiliateReportService) listCommissions(filter repository.AffiliateReportCommissionListFilter, includePromoter bool) ([]AffiliateReportCommission, int64, error) {
	if s == nil || s.reportRepo == nil {
		return []AffiliateReportCommission{}, 0, nil
	}
	rows, total, err := s.reportRepo.ListCommissions(filter)
	if err != nil {
		return nil, 0, err
	}
	buyerMap, err := s.buyerEmailMap(rows)
	if err != nil {
		return nil, 0, err
	}
	result := make([]AffiliateReportCommission, 0, len(rows))
	for i := range rows {
		result = append(result, buildAffiliateReportCommission(&rows[i], buyerMap, includePromoter, filter.ProductID))
	}
	return result, total, nil
}

func (s *AffiliateReportService) buyerEmailMap(rows []models.AffiliateCommission) (map[uint]string, error) {
	userIDs := make([]uint, 0)
	seen := make(map[uint]bool)
	for i := range rows {
		id := rows[i].Order.UserID
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		userIDs = append(userIDs, id)
	}
	result := make(map[uint]string, len(userIDs))
	if len(userIDs) == 0 || s.userRepo == nil {
		return result, nil
	}
	users, err := s.userRepo.ListByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	for i := range users {
		result[users[i].ID] = users[i].Email
	}
	return result, nil
}

func normalizeAffiliateReportFilter(query AffiliateReportQuery) (repository.AffiliateReportFilter, error) {
	now := time.Now()
	start := now.AddDate(0, 0, -affiliateReportDefaultRangeDays)
	end := now
	if query.StartAt != nil {
		start = *query.StartAt
	}
	if query.EndAt != nil {
		end = *query.EndAt
	}
	if !start.Before(end) {
		return repository.AffiliateReportFilter{}, ErrAffiliateReportRangeInvalid
	}
	return repository.AffiliateReportFilter{
		StartAt:            start,
		EndAt:              end,
		AffiliateProfileID: query.AffiliateProfileID,
		AffiliateKeyword:   strings.TrimSpace(query.AffiliateKeyword),
		ProductID:          query.ProductID,
		Status:             strings.TrimSpace(query.Status),
	}, nil
}

func buildAffiliateReportSummary(row repository.AffiliateReportSummaryRow) AffiliateReportSummary {
	totalCommission := row.TotalCommission.Round(2)
	validOrders := row.ValidOrderCount
	avgCommission := decimal.Zero
	if validOrders > 0 {
		avgCommission = totalCommission.Div(decimal.NewFromInt(validOrders)).Round(2)
	}
	avgRate := decimal.Zero
	if row.CommissionBaseAmount.GreaterThan(decimal.Zero) {
		avgRate = totalCommission.Mul(decimal.NewFromInt(100)).Div(row.CommissionBaseAmount).Round(2)
	}
	return AffiliateReportSummary{
		TotalSalesAmount:         moneyText(row.TotalSalesAmount),
		CommissionBaseAmount:     moneyText(row.CommissionBaseAmount),
		TotalCommission:          moneyText(totalCommission),
		AvailableCommission:      moneyText(row.AvailableCommission),
		PendingCommission:        moneyText(row.PendingCommission),
		WithdrawnCommission:      moneyText(row.WithdrawnCommission),
		RejectedCommission:       moneyText(row.RejectedCommission),
		ValidOrderCount:          validOrders,
		ClickCount:               row.ClickCount,
		CustomerCount:            row.CustomerCount,
		NewCustomerOrderCount:    row.NewCustomerOrderCount,
		RepeatCustomerOrderCount: row.RepeatCustomerOrderCount,
		ConversionRate:           percentText(calcAffiliateReportConversion(validOrders, row.ClickCount)),
		AverageCommission:        moneyText(avgCommission),
		AverageCommissionRate:    percentText(avgRate),
	}
}

func emptyAffiliateReportResponse() AffiliateReportResponse {
	return AffiliateReportResponse{
		Summary: buildAffiliateReportSummary(repository.AffiliateReportSummaryRow{}),
		Trend:   []AffiliateReportTrendPoint{},
		SourceBreakdown: []AffiliateReportSourceBreakdown{
			{
				Source:           AffiliateReportSourceProduct,
				Label:            affiliateReportSourceLabel(AffiliateReportSourceProduct),
				SalesAmount:      "0.00",
				CommissionAmount: "0.00",
			},
			{
				Source:           AffiliateReportSourceDefault,
				Label:            affiliateReportSourceLabel(AffiliateReportSourceDefault),
				SalesAmount:      "0.00",
				CommissionAmount: "0.00",
			},
		},
	}
}

func buildAffiliateReportTrend(rows []repository.AffiliateReportTrendRow, startAt, endAt time.Time) []AffiliateReportTrendPoint {
	byDate := make(map[string]decimal.Decimal, len(rows))
	for _, row := range rows {
		byDate[row.Date] = row.CommissionAmount
	}
	loc := startAt.Location()
	if loc == nil {
		loc = time.UTC
	}
	startDay := time.Date(startAt.In(loc).Year(), startAt.In(loc).Month(), startAt.In(loc).Day(), 0, 0, 0, 0, loc)
	endInLoc := endAt.In(loc)
	result := make([]AffiliateReportTrendPoint, 0, len(rows))
	for day := startDay; day.Before(endInLoc); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		amount := byDate[date]
		result = append(result, AffiliateReportTrendPoint{
			Date:             date,
			CommissionAmount: moneyText(amount),
		})
	}
	if len(result) == 0 {
		for _, row := range rows {
			result = append(result, AffiliateReportTrendPoint{
				Date:             row.Date,
				CommissionAmount: moneyText(row.CommissionAmount),
			})
		}
	}
	return result
}

func buildAffiliateReportTopAffiliates(rows []repository.AffiliateReportTopAffiliateRow) []AffiliateReportTopAffiliate {
	result := make([]AffiliateReportTopAffiliate, 0, len(rows))
	for _, row := range rows {
		result = append(result, AffiliateReportTopAffiliate{
			AffiliateProfileID: row.AffiliateProfileID,
			AffiliateCode:      row.AffiliateCode,
			Email:              row.Email,
			DisplayName:        row.DisplayName,
			CommissionAmount:   moneyText(row.CommissionAmount),
			ValidOrderCount:    row.ValidOrderCount,
			ClickCount:         row.ClickCount,
			ConversionRate:     percentText(calcAffiliateReportConversion(row.ValidOrderCount, row.ClickCount)),
		})
	}
	return result
}

func buildAffiliateReportSources(rows []repository.AffiliateReportSourceRow) []AffiliateReportSourceBreakdown {
	result := make([]AffiliateReportSourceBreakdown, 0, len(rows))
	for _, row := range rows {
		result = append(result, AffiliateReportSourceBreakdown{
			Source:           row.Source,
			Label:            affiliateReportSourceLabel(row.Source),
			SalesAmount:      moneyText(row.SalesAmount),
			CommissionAmount: moneyText(row.CommissionAmount),
		})
	}
	return result
}

func buildAffiliateReportCommission(row *models.AffiliateCommission, buyerMap map[uint]string, includePromoter bool, productID uint) AffiliateReportCommission {
	if row == nil {
		return AffiliateReportCommission{}
	}
	source := resolveAffiliateReportCommissionSource(row.Items)
	orderTotalAmount := row.Order.TotalAmount
	baseAmount := row.BaseAmount
	ratePercent := row.RatePercent
	commissionAmount := row.CommissionAmount
	if productID != 0 {
		baseAmount, ratePercent, commissionAmount = summarizeAffiliateReportCommissionItems(row.Items)
		orderTotalAmount = baseAmount
	}
	buyerEmail := strings.TrimSpace(resolveBuyerEmail(row.Order, buyerMap))
	result := AffiliateReportCommission{
		ID:               row.ID,
		OrderID:          row.OrderID,
		OrderNo:          row.Order.OrderNo,
		BuyerEmail:       buyerEmail,
		BuyerMasked:      maskBuyerLabel(buyerEmail),
		OrderTotalAmount: orderTotalAmount,
		BaseAmount:       baseAmount,
		RatePercent:      ratePercent,
		CommissionAmount: commissionAmount,
		SourceType:       source,
		SourceLabel:      affiliateReportSourceLabel(source),
		Status:           row.Status,
		CreatedAt:        row.CreatedAt,
		AvailableAt:      row.AvailableAt,
		CommissionItems:  buildAffiliateReportCommissionItems(row.Items),
	}
	if includePromoter {
		result.AffiliateProfileID = row.AffiliateProfileID
		result.AffiliateCode = row.AffiliateProfile.AffiliateCode
		result.PromoterEmail = row.AffiliateProfile.User.Email
		result.PromoterName = row.AffiliateProfile.User.DisplayName
	}
	return result
}

func summarizeAffiliateReportCommissionItems(items []models.AffiliateCommissionItem) (models.Money, models.Money, models.Money) {
	baseAmount := decimal.Zero
	commissionAmount := decimal.Zero
	for i := range items {
		baseAmount = baseAmount.Add(items[i].BaseAmount.Decimal)
		commissionAmount = commissionAmount.Add(items[i].CommissionAmount.Decimal)
	}
	ratePercent := decimal.Zero
	if baseAmount.GreaterThan(decimal.Zero) {
		ratePercent = commissionAmount.Mul(decimal.NewFromInt(100)).Div(baseAmount).Round(2)
	}
	return models.NewMoneyFromDecimal(baseAmount), models.NewMoneyFromDecimal(ratePercent), models.NewMoneyFromDecimal(commissionAmount)
}

func buildAffiliateReportCommissionItems(items []models.AffiliateCommissionItem) []AffiliateReportCommissionItem {
	result := make([]AffiliateReportCommissionItem, 0, len(items))
	for i := range items {
		source := AffiliateReportSourceDefault
		if items[i].Product.AffiliateCommissionRate != nil {
			source = AffiliateReportSourceProduct
		}
		result = append(result, AffiliateReportCommissionItem{
			ID:               items[i].ID,
			OrderItemID:      items[i].OrderItemID,
			ProductID:        items[i].ProductID,
			ProductTitle:     items[i].OrderItem.TitleJSON,
			SKUSnapshot:      items[i].OrderItem.SKUSnapshotJSON,
			Quantity:         items[i].OrderItem.Quantity,
			BaseAmount:       items[i].BaseAmount,
			RatePercent:      items[i].RatePercent,
			CommissionAmount: items[i].CommissionAmount,
			SourceType:       source,
			SourceLabel:      affiliateReportSourceLabel(source),
		})
	}
	return result
}

func resolveAffiliateReportCommissionSource(items []models.AffiliateCommissionItem) string {
	hasProduct := false
	hasDefault := false
	for i := range items {
		if items[i].Product.AffiliateCommissionRate != nil {
			hasProduct = true
		} else {
			hasDefault = true
		}
	}
	if hasProduct && hasDefault {
		return AffiliateReportSourceMixed
	}
	if hasProduct {
		return AffiliateReportSourceProduct
	}
	if hasDefault {
		return AffiliateReportSourceDefault
	}
	return AffiliateReportSourceDisabled
}

func affiliateReportSourceLabel(source string) string {
	switch source {
	case AffiliateReportSourceProduct:
		return "单商品返佣"
	case AffiliateReportSourceDefault:
		return "平台默认返佣"
	case AffiliateReportSourceDisabled:
		return "不参与返佣"
	case AffiliateReportSourceMixed:
		return "混合返佣"
	default:
		return source
	}
}

func resolveBuyerEmail(order models.Order, buyerMap map[uint]string) string {
	if order.UserID > 0 {
		return buyerMap[order.UserID]
	}
	return order.GuestEmail
}

func maskBuyerLabel(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "已注册用户"
	}
	parts := strings.Split(value, "@")
	if len(parts) == 2 && parts[0] != "" {
		prefix := firstRune(parts[0])
		return prefix + "***@" + parts[1]
	}
	return firstRune(value) + "***"
}

func firstRune(value string) string {
	r, _ := utf8.DecodeRuneInString(value)
	if r == utf8.RuneError {
		return "*"
	}
	return string(r)
}

func calcAffiliateReportConversion(validOrders, clicks int64) decimal.Decimal {
	if clicks <= 0 || validOrders <= 0 {
		return decimal.Zero
	}
	value := (float64(validOrders) / float64(clicks)) * 100
	return decimal.NewFromFloat(math.Round(value*100) / 100).Round(2)
}

func moneyText(value decimal.Decimal) string {
	return value.Round(2).StringFixed(2)
}

func percentText(value decimal.Decimal) string {
	return value.Round(2).StringFixed(2)
}
