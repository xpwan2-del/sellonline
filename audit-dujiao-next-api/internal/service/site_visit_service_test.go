package service

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

type fakeSiteVisitRepo struct {
	rows []repository.SiteVisitTrendRow
}

func (f fakeSiteVisitRepo) Create(*models.SiteVisitEvent) error { return nil }
func (f fakeSiteVisitRepo) GetSummary(repository.SiteVisitFilter) (repository.SiteVisitSummaryRow, error) {
	return repository.SiteVisitSummaryRow{}, nil
}
func (f fakeSiteVisitRepo) GetTrends(repository.SiteVisitFilter) ([]repository.SiteVisitTrendRow, error) {
	return f.rows, nil
}
func (f fakeSiteVisitRepo) GetSources(repository.SiteVisitFilter) ([]repository.SiteVisitSourceRow, error) {
	return nil, nil
}
func (f fakeSiteVisitRepo) GetPages(repository.SiteVisitFilter, int) ([]repository.SiteVisitPageRow, error) {
	return nil, nil
}
func (f fakeSiteVisitRepo) ListRecent(repository.SiteVisitFilter, int, int) ([]models.SiteVisitEvent, int64, error) {
	return nil, 0, nil
}

func TestSiteVisitServiceTrendFillsEmptyDays(t *testing.T) {
	service := NewSiteVisitService(fakeSiteVisitRepo{
		rows: []repository.SiteVisitTrendRow{
			{Day: "2026-06-05", PV: 3, UV: 2},
		},
	})
	from := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC)
	result, err := service.GetTrends(SiteVisitQueryInput{
		Range:    "custom",
		From:     &from,
		To:       &to,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("GetTrends: %v", err)
	}
	if len(result.Points) != 3 {
		t.Fatalf("points len = %d, want 3", len(result.Points))
	}
	if result.Points[0].Date != "2026-06-04" || result.Points[0].PV != 0 {
		t.Fatalf("first point = %+v, want empty 2026-06-04", result.Points[0])
	}
	if result.Points[1].Date != "2026-06-05" || result.Points[1].PV != 3 {
		t.Fatalf("second point = %+v, want pv 3 on 2026-06-05", result.Points[1])
	}
	if result.Points[2].Date != "2026-06-06" || result.Points[2].PV != 0 {
		t.Fatalf("third point = %+v, want empty 2026-06-06", result.Points[2])
	}
}

func TestNormalizeVisitSourcePrefersAffiliateCode(t *testing.T) {
	if got := normalizeVisitSource("", "ABC123", ""); got != SiteVisitSourceAffiliate {
		t.Fatalf("source = %s, want affiliate", got)
	}
	if got := normalizeVisitSource("", "", "https://www.google.com/search?q=test"); got != SiteVisitSourceSearch {
		t.Fatalf("source = %s, want search", got)
	}
	if got := normalizeVisitSource("", "", ""); got != SiteVisitSourceDirect {
		t.Fatalf("source = %s, want direct", got)
	}
}
