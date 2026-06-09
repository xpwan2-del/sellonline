package repository

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAffiliateReportListCommissionsSQLiteDistinct(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.AffiliateProfile{},
		&models.Order{},
		&models.OrderItem{},
		&models.Product{},
		&models.AffiliateCommission{},
		&models.AffiliateCommissionItem{},
		&models.AffiliateClick{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	paidAt := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	money := func(value int64) models.Money {
		return models.NewMoneyFromDecimal(decimal.NewFromInt(value))
	}
	rate := money(10)
	product := models.Product{
		CategoryID:              1,
		Slug:                    "report-product",
		TitleJSON:               models.JSON{"zh-CN": "报表商品"},
		PriceAmount:             money(100),
		AffiliateCommissionRate: &rate,
		IsAffiliateEnabled:      true,
		IsActive:                true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	productB := models.Product{
		CategoryID:              1,
		Slug:                    "report-product-b",
		TitleJSON:               models.JSON{"zh-CN": "报表商品 B"},
		PriceAmount:             money(100),
		AffiliateCommissionRate: &rate,
		IsAffiliateEnabled:      true,
		IsActive:                true,
	}
	if err := db.Create(&productB).Error; err != nil {
		t.Fatalf("create product b: %v", err)
	}
	promoter := models.User{Email: "promoter@example.test", PasswordHash: "hash", DisplayName: "Promoter"}
	buyer := models.User{Email: "buyer@example.test", PasswordHash: "hash", DisplayName: "Buyer"}
	if err := db.Create(&promoter).Error; err != nil {
		t.Fatalf("create promoter: %v", err)
	}
	if err := db.Create(&buyer).Error; err != nil {
		t.Fatalf("create buyer: %v", err)
	}
	profile := models.AffiliateProfile{
		UserID:        promoter.ID,
		AffiliateCode: "REPORT1",
		Status:        "active",
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	order := models.Order{
		OrderNo:            "DJ202606060001",
		UserID:             buyer.ID,
		Status:             constants.OrderStatusPaid,
		Currency:           "USD",
		TotalAmount:        money(300),
		AffiliateProfileID: &profile.ID,
		AffiliateCode:      profile.AffiliateCode,
		PaidAt:             &paidAt,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	items := []models.OrderItem{
		{OrderID: order.ID, ProductID: product.ID, TitleJSON: models.JSON{"zh-CN": "报表商品 A"}, UnitPrice: money(100), Quantity: 1, TotalPrice: money(100), FulfillmentType: "manual"},
		{OrderID: order.ID, ProductID: product.ID, TitleJSON: models.JSON{"zh-CN": "报表商品 B"}, UnitPrice: money(100), Quantity: 1, TotalPrice: money(100), FulfillmentType: "manual"},
		{OrderID: order.ID, ProductID: productB.ID, TitleJSON: models.JSON{"zh-CN": "报表商品 C"}, UnitPrice: money(100), Quantity: 1, TotalPrice: money(100), FulfillmentType: "manual"},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("create order items: %v", err)
	}
	commission := models.AffiliateCommission{
		AffiliateProfileID: profile.ID,
		OrderID:            order.ID,
		CommissionType:     "order",
		BaseAmount:         money(300),
		RatePercent:        money(10),
		CommissionAmount:   money(30),
		Status:             constants.AffiliateCommissionStatusAvailable,
	}
	if err := db.Create(&commission).Error; err != nil {
		t.Fatalf("create commission: %v", err)
	}
	commissionItems := []models.AffiliateCommissionItem{
		{AffiliateCommissionID: commission.ID, OrderItemID: items[0].ID, ProductID: product.ID, BaseAmount: money(100), RatePercent: money(10), CommissionAmount: money(10)},
		{AffiliateCommissionID: commission.ID, OrderItemID: items[1].ID, ProductID: product.ID, BaseAmount: money(100), RatePercent: money(10), CommissionAmount: money(10)},
		{AffiliateCommissionID: commission.ID, OrderItemID: items[2].ID, ProductID: productB.ID, BaseAmount: money(100), RatePercent: money(10), CommissionAmount: money(10)},
	}
	if err := db.Create(&commissionItems).Error; err != nil {
		t.Fatalf("create commission items: %v", err)
	}

	repo := NewAffiliateReportRepository(db)
	rows, total, err := repo.ListCommissions(AffiliateReportCommissionListFilter{
		AffiliateReportFilter: AffiliateReportFilter{
			StartAt:   paidAt.Add(-time.Hour),
			EndAt:     paidAt.Add(time.Hour),
			ProductID: product.ID,
		},
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list commissions: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(rows) != 1 {
		t.Fatalf("rows len = %d, want 1", len(rows))
	}
	if len(rows[0].Items) != 2 {
		t.Fatalf("preloaded items len = %d, want 2", len(rows[0].Items))
	}
	summary, err := repo.GetSummary(AffiliateReportFilter{
		StartAt:   paidAt.Add(-time.Hour),
		EndAt:     paidAt.Add(time.Hour),
		ProductID: product.ID,
	})
	if err != nil {
		t.Fatalf("get summary by product: %v", err)
	}
	if !summary.CommissionBaseAmount.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("summary base = %s, want 200.00", summary.CommissionBaseAmount)
	}
	if !summary.TotalCommission.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("summary commission = %s, want 20.00", summary.TotalCommission)
	}
	if summary.ValidOrderCount != 1 {
		t.Fatalf("summary valid orders = %d, want 1", summary.ValidOrderCount)
	}

	rows, total, err = repo.ListCommissions(AffiliateReportCommissionListFilter{
		AffiliateReportFilter: AffiliateReportFilter{
			StartAt:          paidAt.Add(-time.Hour),
			EndAt:            paidAt.Add(time.Hour),
			AffiliateKeyword: "promoter@example.test",
		},
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list commissions by affiliate keyword: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("keyword filtered result total=%d rows=%d, want 1/1", total, len(rows))
	}
}
