package repository

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"

	"gorm.io/gorm"
)

// SiteVisitRepository 进站统计数据访问接口。
type SiteVisitRepository interface {
	Create(event *models.SiteVisitEvent) error
	GetSummary(filter SiteVisitFilter) (SiteVisitSummaryRow, error)
	GetTrends(filter SiteVisitFilter) ([]SiteVisitTrendRow, error)
	GetSources(filter SiteVisitFilter) ([]SiteVisitSourceRow, error)
	GetPages(filter SiteVisitFilter, limit int) ([]SiteVisitPageRow, error)
	ListRecent(filter SiteVisitFilter, page, pageSize int) ([]models.SiteVisitEvent, int64, error)
}

// SiteVisitFilter 进站统计查询条件。
type SiteVisitFilter struct {
	StartAt    time.Time
	EndAt      time.Time
	Timezone   string
	SourceType string
	DeviceType string
	PageType   string
}

// SiteVisitSummaryRow 进站统计汇总。
type SiteVisitSummaryRow struct {
	PV          int64
	UV          int64
	UniqueIP    int64
	AffiliatePV int64
	LoggedInPV  int64
	VisitorPV   int64
	MobilePV    int64
	DesktopPV   int64
	TabletPV    int64
	DirectPV    int64
	SearchPV    int64
	ExternalPV  int64
	UnknownPV   int64
}

// SiteVisitTrendRow 按天趋势。
type SiteVisitTrendRow struct {
	Day         string
	PV          int64
	UV          int64
	AffiliatePV int64
	DirectPV    int64
	SearchPV    int64
	ExternalPV  int64
	UnknownPV   int64
}

// SiteVisitSourceRow 来源统计。
type SiteVisitSourceRow struct {
	SourceType string
	PV         int64
	UV         int64
}

// SiteVisitPageRow 页面排行。
type SiteVisitPageRow struct {
	Path string
	PV   int64
	UV   int64
}

// GormSiteVisitRepository GORM 实现。
type GormSiteVisitRepository struct {
	db *gorm.DB
}

// NewSiteVisitRepository 创建进站统计仓库。
func NewSiteVisitRepository(db *gorm.DB) *GormSiteVisitRepository {
	return &GormSiteVisitRepository{db: db}
}

func (r *GormSiteVisitRepository) Create(event *models.SiteVisitEvent) error {
	if event == nil {
		return nil
	}
	return r.db.Create(event).Error
}

func (r *GormSiteVisitRepository) GetSummary(filter SiteVisitFilter) (SiteVisitSummaryRow, error) {
	var row SiteVisitSummaryRow
	query := r.applyFilter(r.db.Model(&models.SiteVisitEvent{}), filter)
	selectSQL := `
		COUNT(*) AS pv,
		COUNT(DISTINCT visitor_key) AS uv,
		COUNT(DISTINCT client_ip) AS unique_ip,
		COALESCE(SUM(CASE WHEN source_type = 'affiliate' THEN 1 ELSE 0 END), 0) AS affiliate_pv,
		COALESCE(SUM(CASE WHEN user_id > 0 THEN 1 ELSE 0 END), 0) AS logged_in_pv,
		COALESCE(SUM(CASE WHEN user_id = 0 THEN 1 ELSE 0 END), 0) AS visitor_pv,
		COALESCE(SUM(CASE WHEN device_type = 'mobile' THEN 1 ELSE 0 END), 0) AS mobile_pv,
		COALESCE(SUM(CASE WHEN device_type = 'desktop' THEN 1 ELSE 0 END), 0) AS desktop_pv,
		COALESCE(SUM(CASE WHEN device_type = 'tablet' THEN 1 ELSE 0 END), 0) AS tablet_pv,
		COALESCE(SUM(CASE WHEN source_type = 'direct' THEN 1 ELSE 0 END), 0) AS direct_pv,
		COALESCE(SUM(CASE WHEN source_type = 'search' THEN 1 ELSE 0 END), 0) AS search_pv,
		COALESCE(SUM(CASE WHEN source_type = 'external' THEN 1 ELSE 0 END), 0) AS external_pv,
		COALESCE(SUM(CASE WHEN source_type = 'unknown' THEN 1 ELSE 0 END), 0) AS unknown_pv
	`
	if err := query.Select(selectSQL).Scan(&row).Error; err != nil {
		return SiteVisitSummaryRow{}, err
	}
	return row, nil
}

func (r *GormSiteVisitRepository) GetTrends(filter SiteVisitFilter) ([]SiteVisitTrendRow, error) {
	loc := loadLocationOrUTC(filter.Timezone)
	dayExpr := dateGroupExpr(r.db, "created_at", loc, filter.StartAt)
	selectSQL := dayExpr + ` AS day,
		COUNT(*) AS pv,
		COUNT(DISTINCT visitor_key) AS uv,
		COALESCE(SUM(CASE WHEN source_type = 'affiliate' THEN 1 ELSE 0 END), 0) AS affiliate_pv,
		COALESCE(SUM(CASE WHEN source_type = 'direct' THEN 1 ELSE 0 END), 0) AS direct_pv,
		COALESCE(SUM(CASE WHEN source_type = 'search' THEN 1 ELSE 0 END), 0) AS search_pv,
		COALESCE(SUM(CASE WHEN source_type = 'external' THEN 1 ELSE 0 END), 0) AS external_pv,
		COALESCE(SUM(CASE WHEN source_type = 'unknown' THEN 1 ELSE 0 END), 0) AS unknown_pv`

	var rows []SiteVisitTrendRow
	err := r.applyFilter(r.db.Model(&models.SiteVisitEvent{}), filter).
		Select(selectSQL).
		Group(dayExpr).
		Order("day asc").
		Scan(&rows).Error
	return rows, err
}

func (r *GormSiteVisitRepository) GetSources(filter SiteVisitFilter) ([]SiteVisitSourceRow, error) {
	var rows []SiteVisitSourceRow
	err := r.applyFilter(r.db.Model(&models.SiteVisitEvent{}), filter).
		Select("source_type, COUNT(*) AS pv, COUNT(DISTINCT visitor_key) AS uv").
		Group("source_type").
		Order("pv desc").
		Scan(&rows).Error
	return rows, err
}

func (r *GormSiteVisitRepository) GetPages(filter SiteVisitFilter, limit int) ([]SiteVisitPageRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var rows []SiteVisitPageRow
	err := r.applyFilter(r.db.Model(&models.SiteVisitEvent{}), filter).
		Select("path, COUNT(*) AS pv, COUNT(DISTINCT visitor_key) AS uv").
		Group("path").
		Order("pv desc").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *GormSiteVisitRepository) ListRecent(filter SiteVisitFilter, page, pageSize int) ([]models.SiteVisitEvent, int64, error) {
	query := r.applyFilter(r.db.Model(&models.SiteVisitEvent{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.SiteVisitEvent
	err := applyPagination(query, page, pageSize).Order("id desc").Find(&rows).Error
	return rows, total, err
}

func (r *GormSiteVisitRepository) applyFilter(query *gorm.DB, filter SiteVisitFilter) *gorm.DB {
	if query == nil {
		return query
	}
	if !filter.StartAt.IsZero() {
		query = query.Where("created_at >= ?", filter.StartAt)
	}
	if !filter.EndAt.IsZero() {
		query = query.Where("created_at < ?", filter.EndAt)
	}
	if filter.SourceType != "" && filter.SourceType != "all" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	if filter.DeviceType != "" && filter.DeviceType != "all" {
		query = query.Where("device_type = ?", filter.DeviceType)
	}
	if filter.PageType != "" && filter.PageType != "all" {
		query = query.Where("page_type = ?", filter.PageType)
	}
	return query
}

func loadLocationOrUTC(name string) *time.Location {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "UTC"
	}
	loc, err := time.LoadLocation(trimmed)
	if err != nil {
		return time.UTC
	}
	return loc
}
