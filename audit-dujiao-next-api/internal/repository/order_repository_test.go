package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func setupOrderRepositoryTest(t *testing.T) (*GormOrderRepository, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:order_repo_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return NewOrderRepository(db), db
}

func TestOrderRepositoryStatsByUser(t *testing.T) {
	repo, db := setupOrderRepositoryTest(t)
	now := time.Now().UTC().Truncate(time.Second)
	money := models.NewMoneyFromDecimal(decimal.RequireFromString("10.00"))

	orders := []models.Order{
		{
			OrderNo:        "ORD-A-PENDING",
			UserID:         1,
			Status:         constants.OrderStatusPendingPayment,
			Currency:       "CNY",
			OriginalAmount: money,
			TotalAmount:    money,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			OrderNo:        "ORD-A-DELIVERED",
			UserID:         1,
			Status:         constants.OrderStatusDelivered,
			Currency:       "CNY",
			OriginalAmount: money,
			TotalAmount:    money,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			OrderNo:        "ORD-B-PENDING",
			UserID:         2,
			Status:         constants.OrderStatusPendingPayment,
			Currency:       "CNY",
			OriginalAmount: money,
			TotalAmount:    money,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			OrderNo:        "OTHER-A-COMPLETED",
			UserID:         1,
			Status:         constants.OrderStatusCompleted,
			Currency:       "CNY",
			OriginalAmount: money,
			TotalAmount:    money,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("create orders failed: %v", err)
	}

	parentID := orders[0].ID
	child := models.Order{
		OrderNo:        "ORD-A-CHILD",
		ParentID:       &parentID,
		UserID:         1,
		Status:         constants.OrderStatusPendingPayment,
		Currency:       "CNY",
		OriginalAmount: money,
		TotalAmount:    money,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("create child order failed: %v", err)
	}

	stats, err := repo.StatsByUser(OrderListFilter{
		UserID: 1,
		Status: constants.OrderStatusPendingPayment,
	})
	if err != nil {
		t.Fatalf("stats by user failed: %v", err)
	}
	if stats[constants.OrderStatusPendingPayment] != 1 {
		t.Fatalf("pending count want 1 got %d", stats[constants.OrderStatusPendingPayment])
	}
	if stats[constants.OrderStatusDelivered] != 1 {
		t.Fatalf("delivered count want 1 got %d", stats[constants.OrderStatusDelivered])
	}
	if stats[constants.OrderStatusCompleted] != 1 {
		t.Fatalf("completed count want 1 got %d", stats[constants.OrderStatusCompleted])
	}

	filteredStats, err := repo.StatsByUser(OrderListFilter{
		UserID:  1,
		OrderNo: "ORD-A",
	})
	if err != nil {
		t.Fatalf("stats by user order_no failed: %v", err)
	}
	if filteredStats[constants.OrderStatusPendingPayment] != 1 {
		t.Fatalf("filtered pending count want 1 got %d", filteredStats[constants.OrderStatusPendingPayment])
	}
	if filteredStats[constants.OrderStatusDelivered] != 1 {
		t.Fatalf("filtered delivered count want 1 got %d", filteredStats[constants.OrderStatusDelivered])
	}
	if filteredStats[constants.OrderStatusCompleted] != 0 {
		t.Fatalf("filtered completed count want 0 got %d", filteredStats[constants.OrderStatusCompleted])
	}
}
