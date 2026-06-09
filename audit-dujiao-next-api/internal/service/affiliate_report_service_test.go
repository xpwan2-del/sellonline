package service

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/shopspring/decimal"
)

type panicAffiliateReportRepo struct{}

func (panicAffiliateReportRepo) GetSummary(repository.AffiliateReportFilter) (repository.AffiliateReportSummaryRow, error) {
	panic("report repository should not be called")
}

func (panicAffiliateReportRepo) GetCustomerStats(repository.AffiliateReportFilter) (repository.AffiliateReportCustomerStatsRow, error) {
	panic("report repository should not be called")
}

func (panicAffiliateReportRepo) GetTrend(repository.AffiliateReportFilter) ([]repository.AffiliateReportTrendRow, error) {
	panic("report repository should not be called")
}

func (panicAffiliateReportRepo) GetTopAffiliates(repository.AffiliateReportFilter, int) ([]repository.AffiliateReportTopAffiliateRow, error) {
	panic("report repository should not be called")
}

func (panicAffiliateReportRepo) GetSourceBreakdown(repository.AffiliateReportFilter) ([]repository.AffiliateReportSourceRow, error) {
	panic("report repository should not be called")
}

func (panicAffiliateReportRepo) ListCommissions(repository.AffiliateReportCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	panic("report repository should not be called")
}

type affiliateReportRepoStub struct {
	trendRows []repository.AffiliateReportTrendRow
}

func (s affiliateReportRepoStub) GetSummary(repository.AffiliateReportFilter) (repository.AffiliateReportSummaryRow, error) {
	return repository.AffiliateReportSummaryRow{}, nil
}

func (s affiliateReportRepoStub) GetCustomerStats(repository.AffiliateReportFilter) (repository.AffiliateReportCustomerStatsRow, error) {
	return repository.AffiliateReportCustomerStatsRow{}, nil
}

func (s affiliateReportRepoStub) GetTrend(repository.AffiliateReportFilter) ([]repository.AffiliateReportTrendRow, error) {
	return s.trendRows, nil
}

func (s affiliateReportRepoStub) GetTopAffiliates(repository.AffiliateReportFilter, int) ([]repository.AffiliateReportTopAffiliateRow, error) {
	return []repository.AffiliateReportTopAffiliateRow{}, nil
}

func (s affiliateReportRepoStub) GetSourceBreakdown(repository.AffiliateReportFilter) ([]repository.AffiliateReportSourceRow, error) {
	return []repository.AffiliateReportSourceRow{}, nil
}

func (s affiliateReportRepoStub) ListCommissions(repository.AffiliateReportCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	return []models.AffiliateCommission{}, 0, nil
}

func TestAffiliateReportServiceUserWithoutProfileReturnsEmptyReport(t *testing.T) {
	svc := NewAffiliateReportService(panicAffiliateReportRepo{}, nil, nil)

	resp, err := svc.GetUserSummary(1001, AffiliateReportQuery{}, "https://example.test")
	if err != nil {
		t.Fatalf("GetUserSummary returned error: %v", err)
	}
	if resp.Summary.TotalCommission != "0.00" {
		t.Fatalf("expected zero total commission, got %q", resp.Summary.TotalCommission)
	}
	if resp.Summary.PromotionURL != "" {
		t.Fatalf("expected empty promotion url, got %q", resp.Summary.PromotionURL)
	}

	rows, total, err := svc.ListUserCommissions(1001, AffiliateReportCommissionQuery{})
	if err != nil {
		t.Fatalf("ListUserCommissions returned error: %v", err)
	}
	if total != 0 || len(rows) != 0 {
		t.Fatalf("expected empty commission list, got total=%d len=%d", total, len(rows))
	}
}

func TestAffiliateReportServiceFillsEmptyTrendDays(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	startAt := time.Date(2026, 6, 1, 0, 0, 0, 0, loc)
	endAt := time.Date(2026, 6, 7, 0, 0, 0, 0, loc)
	svc := NewAffiliateReportService(affiliateReportRepoStub{
		trendRows: []repository.AffiliateReportTrendRow{
			{Date: "2026-06-06", CommissionAmount: decimal.NewFromInt(1530)},
		},
	}, nil, nil)

	resp, err := svc.GetAdminSummary(AffiliateReportQuery{
		StartAt: &startAt,
		EndAt:   &endAt,
	})
	if err != nil {
		t.Fatalf("GetAdminSummary returned error: %v", err)
	}
	if len(resp.Trend) != 6 {
		t.Fatalf("trend len = %d, want 6", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-06-01" || resp.Trend[0].CommissionAmount != "0.00" {
		t.Fatalf("unexpected first trend point: %+v", resp.Trend[0])
	}
	last := resp.Trend[len(resp.Trend)-1]
	if last.Date != "2026-06-06" || last.CommissionAmount != "1530.00" {
		t.Fatalf("unexpected last trend point: %+v", last)
	}
}

func TestAffiliateReportCommissionUsesFilteredItemAmounts(t *testing.T) {
	money := func(value int64) models.Money {
		return models.NewMoneyFromDecimal(decimal.NewFromInt(value))
	}
	productRate := money(15)
	row := models.AffiliateCommission{
		ID:               10,
		OrderID:          20,
		BaseAmount:       money(1000),
		RatePercent:      money(20),
		CommissionAmount: money(200),
		Status:           "available",
		Order: models.Order{
			ID:          20,
			OrderNo:     "DJ202606070001",
			TotalAmount: money(1000),
			UserID:      30,
		},
		Items: []models.AffiliateCommissionItem{
			{
				ID:               1,
				ProductID:        100,
				BaseAmount:       money(300),
				RatePercent:      money(15),
				CommissionAmount: money(45),
				Product: models.Product{
					AffiliateCommissionRate: &productRate,
				},
				OrderItem: models.OrderItem{
					ProductID: 100,
					Quantity:  3,
				},
			},
		},
	}

	resp := buildAffiliateReportCommission(&row, map[uint]string{30: "buyer@example.test"}, false, 100)
	if resp.OrderTotalAmount.String() != "300.00" {
		t.Fatalf("order total amount = %s, want 300.00", resp.OrderTotalAmount.String())
	}
	if resp.BaseAmount.String() != "300.00" {
		t.Fatalf("base amount = %s, want 300.00", resp.BaseAmount.String())
	}
	if resp.RatePercent.String() != "15.00" {
		t.Fatalf("rate percent = %s, want 15.00", resp.RatePercent.String())
	}
	if resp.CommissionAmount.String() != "45.00" {
		t.Fatalf("commission amount = %s, want 45.00", resp.CommissionAmount.String())
	}
	if len(resp.CommissionItems) != 1 || resp.CommissionItems[0].ProductID != 100 {
		t.Fatalf("unexpected commission items: %+v", resp.CommissionItems)
	}
	if resp.SourceType != AffiliateReportSourceProduct {
		t.Fatalf("source type = %s, want %s", resp.SourceType, AffiliateReportSourceProduct)
	}
}
