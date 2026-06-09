package public

import (
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

func TestDecorateProductStock_AutoSkipsInactiveSKUs(t *testing.T) {
	h := &Handler{}
	product := &models.Product{
		ID:              1,
		FulfillmentType: constants.FulfillmentTypeAuto,
		SKUs: []models.ProductSKU{
			{
				ID:                 11,
				SKUCode:            models.DefaultSKUCode,
				IsActive:           true,
				AutoStockAvailable: 2,
				AutoStockTotal:     3,
				AutoStockLocked:    1,
				AutoStockSold:      4,
			},
			{
				ID:                 12,
				SKUCode:            "DISABLED",
				IsActive:           false,
				AutoStockAvailable: 100,
				AutoStockTotal:     120,
				AutoStockLocked:    20,
				AutoStockSold:      50,
			},
		},
	}

	item := publicProductView{Product: *product}
	h.decorateProductStock(product, &item)

	if item.AutoStockAvailable != 2 {
		t.Fatalf("expected auto_stock_available=2, got %d", item.AutoStockAvailable)
	}
	if item.AutoStockTotal != 3 {
		t.Fatalf("expected auto_stock_total=3, got %d", item.AutoStockTotal)
	}
	if item.AutoStockLocked != 1 {
		t.Fatalf("expected auto_stock_locked=1, got %d", item.AutoStockLocked)
	}
	if item.AutoStockSold != 4 {
		t.Fatalf("expected auto_stock_sold=4, got %d", item.AutoStockSold)
	}
	if item.IsSoldOut {
		t.Fatalf("expected product not sold out when active sku has stock")
	}
}

func TestDecorateProductStock_AutoSoldCountUsesActiveSKUsAndOffset(t *testing.T) {
	h := &Handler{}
	product := &models.Product{
		ID:              1,
		FulfillmentType: constants.FulfillmentTypeAuto,
		SoldCountOffset: 6,
		SKUs: []models.ProductSKU{
			{
				ID:                 11,
				SKUCode:            models.DefaultSKUCode,
				IsActive:           true,
				AutoStockAvailable: 2,
				AutoStockSold:      4,
			},
			{
				ID:                 12,
				SKUCode:            "DISABLED",
				IsActive:           false,
				AutoStockAvailable: 10,
				AutoStockSold:      50,
			},
		},
	}

	item := publicProductView{Product: *product}
	h.decorateProductStock(product, &item)

	if item.RealSoldCount != 4 {
		t.Fatalf("expected real_sold_count=4, got %d", item.RealSoldCount)
	}
	if item.SoldCountOffset != 6 {
		t.Fatalf("expected sold_count_offset=6, got %d", item.SoldCountOffset)
	}
	if item.SoldCount != 10 {
		t.Fatalf("expected sold_count=10, got %d", item.SoldCount)
	}

	resp := item.toProductResp()
	if resp.RealSoldCount != 4 || resp.SoldCountOffset != 6 || resp.SoldCount != 10 {
		t.Fatalf("unexpected response sold counts: real=%d offset=%d sold=%d", resp.RealSoldCount, resp.SoldCountOffset, resp.SoldCount)
	}
}

func TestDecorateProductStock_ManualSoldCountUsesActiveSKUsAndNormalizesOffset(t *testing.T) {
	h := &Handler{}
	product := &models.Product{
		ID:              1,
		FulfillmentType: constants.FulfillmentTypeManual,
		SoldCountOffset: -3,
		SKUs: []models.ProductSKU{
			{
				ID:               11,
				SKUCode:          models.DefaultSKUCode,
				IsActive:         true,
				ManualStockTotal: 2,
				ManualStockSold:  5,
			},
			{
				ID:               12,
				SKUCode:          "PRO",
				IsActive:         true,
				ManualStockTotal: 3,
				ManualStockSold:  7,
			},
			{
				ID:               13,
				SKUCode:          "DISABLED",
				IsActive:         false,
				ManualStockTotal: 100,
				ManualStockSold:  50,
			},
		},
	}

	item := publicProductView{Product: *product}
	h.decorateProductStock(product, &item)

	if item.RealSoldCount != 12 {
		t.Fatalf("expected real_sold_count=12, got %d", item.RealSoldCount)
	}
	if item.SoldCountOffset != 0 {
		t.Fatalf("expected negative sold_count_offset to normalize to 0, got %d", item.SoldCountOffset)
	}
	if item.SoldCount != 12 {
		t.Fatalf("expected sold_count=12, got %d", item.SoldCount)
	}
}
