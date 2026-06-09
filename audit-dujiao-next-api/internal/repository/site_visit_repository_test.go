package repository

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSiteVisitRepositoryTest(t *testing.T) (*GormSiteVisitRepository, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&models.SiteVisitEvent{}); err != nil {
		t.Fatalf("migrate site visit failed: %v", err)
	}
	return NewSiteVisitRepository(db), db
}

func TestSiteVisitRepositorySummaryAndBreakdowns(t *testing.T) {
	repo, db := setupSiteVisitRepositoryTest(t)
	now := time.Date(2026, 6, 6, 8, 0, 0, 0, time.UTC)
	events := []models.SiteVisitEvent{
		{VisitorKey: "v1", SessionKey: "s1", Path: "/", SourceType: "direct", ClientIP: "10.0.0.1", DeviceType: "desktop", CreatedAt: now},
		{VisitorKey: "v1", SessionKey: "s1", Path: "/products/a", SourceType: "affiliate", AffiliateCode: "ABC123", ClientIP: "10.0.0.1", DeviceType: "desktop", CreatedAt: now.Add(time.Hour)},
		{VisitorKey: "v2", SessionKey: "s2", UserID: 9, Path: "/cart", SourceType: "search", ClientIP: "10.0.0.2", DeviceType: "mobile", CreatedAt: now.Add(2 * time.Hour)},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("create events failed: %v", err)
	}

	filter := SiteVisitFilter{
		StartAt:  now.Add(-time.Hour),
		EndAt:    now.Add(24 * time.Hour),
		Timezone: "Asia/Shanghai",
	}
	summary, err := repo.GetSummary(filter)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.PV != 3 || summary.UV != 2 || summary.UniqueIP != 2 {
		t.Fatalf("summary = pv %d uv %d ip %d, want 3/2/2", summary.PV, summary.UV, summary.UniqueIP)
	}
	if summary.AffiliatePV != 1 || summary.SearchPV != 1 || summary.DirectPV != 1 {
		t.Fatalf("unexpected source summary: %+v", summary)
	}
	if summary.LoggedInPV != 1 || summary.MobilePV != 1 || summary.DesktopPV != 2 {
		t.Fatalf("unexpected user/device summary: %+v", summary)
	}

	sources, err := repo.GetSources(filter)
	if err != nil {
		t.Fatalf("GetSources: %v", err)
	}
	if len(sources) != 3 {
		t.Fatalf("sources len = %d, want 3", len(sources))
	}

	pages, err := repo.GetPages(filter, 10)
	if err != nil {
		t.Fatalf("GetPages: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("pages len = %d, want 3", len(pages))
	}
}

func TestSiteVisitRepositoryTrendGroupsByTimezoneDay(t *testing.T) {
	repo, db := setupSiteVisitRepositoryTest(t)
	events := []models.SiteVisitEvent{
		{VisitorKey: "v1", Path: "/", SourceType: "direct", ClientIP: "10.0.0.1", DeviceType: "desktop", CreatedAt: time.Date(2026, 6, 5, 16, 30, 0, 0, time.UTC)},
		{VisitorKey: "v2", Path: "/", SourceType: "affiliate", ClientIP: "10.0.0.2", DeviceType: "mobile", CreatedAt: time.Date(2026, 6, 6, 1, 0, 0, 0, time.UTC)},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("create events failed: %v", err)
	}
	rows, err := repo.GetTrends(SiteVisitFilter{
		StartAt:  time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
		EndAt:    time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC),
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("GetTrends: %v", err)
	}
	if len(rows) != 1 || rows[0].Day != "2026-06-06" || rows[0].PV != 2 {
		t.Fatalf("trend rows = %+v, want one 2026-06-06 row with pv 2", rows)
	}
}
