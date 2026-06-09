package service

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

const (
	SiteVisitSourceDirect    = "direct"
	SiteVisitSourceAffiliate = "affiliate"
	SiteVisitSourceSearch    = "search"
	SiteVisitSourceExternal  = "external"
	SiteVisitSourceUnknown   = "unknown"

	SiteVisitDeviceDesktop = "desktop"
	SiteVisitDeviceMobile  = "mobile"
	SiteVisitDeviceTablet  = "tablet"
	SiteVisitDeviceUnknown = "unknown"

	siteVisitCustomMaxDays = 120
)

var ErrSiteVisitRangeInvalid = errors.New("site visit range invalid")

// SiteVisitService 进站统计服务。
type SiteVisitService struct {
	repo repository.SiteVisitRepository
}

// NewSiteVisitService 创建进站统计服务。
func NewSiteVisitService(repo repository.SiteVisitRepository) *SiteVisitService {
	return &SiteVisitService{repo: repo}
}

// RecordSiteVisitInput 记录访问输入。
type RecordSiteVisitInput struct {
	VisitorKey    string
	SessionKey    string
	UserID        uint
	Path          string
	PageType      string
	SourceType    string
	Referrer      string
	AffiliateCode string
	ClientIP      string
	UserAgent     string
	DeviceType    string
}

// SiteVisitQueryInput 查询输入。
type SiteVisitQueryInput struct {
	Range      string
	From       *time.Time
	To         *time.Time
	Timezone   string
	SourceType string
	DeviceType string
	PageType   string
	Page       int
	PageSize   int
}

// SiteVisitSummaryResponse 汇总响应。
type SiteVisitSummaryResponse struct {
	Range    string              `json:"range"`
	From     string              `json:"from"`
	To       string              `json:"to"`
	Timezone string              `json:"timezone"`
	KPI      SiteVisitSummaryKPI `json:"kpi"`
}

// SiteVisitSummaryKPI 汇总指标。
type SiteVisitSummaryKPI struct {
	PV            int64  `json:"pv"`
	UV            int64  `json:"uv"`
	UniqueIP      int64  `json:"unique_ip"`
	AffiliatePV   int64  `json:"affiliate_pv"`
	LoggedInPV    int64  `json:"logged_in_pv"`
	VisitorPV     int64  `json:"visitor_pv"`
	MobilePV      int64  `json:"mobile_pv"`
	DesktopPV     int64  `json:"desktop_pv"`
	TabletPV      int64  `json:"tablet_pv"`
	DirectPV      int64  `json:"direct_pv"`
	SearchPV      int64  `json:"search_pv"`
	ExternalPV    int64  `json:"external_pv"`
	UnknownPV     int64  `json:"unknown_pv"`
	AffiliateRate string `json:"affiliate_rate"`
	LoggedInRate  string `json:"logged_in_rate"`
	MobileRate    string `json:"mobile_rate"`
}

// SiteVisitTrendResponse 趋势响应。
type SiteVisitTrendResponse struct {
	Range    string                `json:"range"`
	From     string                `json:"from"`
	To       string                `json:"to"`
	Timezone string                `json:"timezone"`
	Points   []SiteVisitTrendPoint `json:"points"`
}

// SiteVisitTrendPoint 趋势点。
type SiteVisitTrendPoint struct {
	Date        string `json:"date"`
	PV          int64  `json:"pv"`
	UV          int64  `json:"uv"`
	AffiliatePV int64  `json:"affiliate_pv"`
	DirectPV    int64  `json:"direct_pv"`
	SearchPV    int64  `json:"search_pv"`
	ExternalPV  int64  `json:"external_pv"`
	UnknownPV   int64  `json:"unknown_pv"`
}

// SiteVisitSourceItem 来源项。
type SiteVisitSourceItem struct {
	SourceType string `json:"source_type"`
	PV         int64  `json:"pv"`
	UV         int64  `json:"uv"`
	Rate       string `json:"rate"`
}

// SiteVisitPageItem 页面排行项。
type SiteVisitPageItem struct {
	Path string `json:"path"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
	Rate string `json:"rate"`
}

// SiteVisitRecentItem 最近访问项。
type SiteVisitRecentItem struct {
	ID            uint   `json:"id"`
	VisitorKey    string `json:"visitor_key"`
	SessionKey    string `json:"session_key"`
	UserID        uint   `json:"user_id"`
	Path          string `json:"path"`
	PageType      string `json:"page_type"`
	SourceType    string `json:"source_type"`
	Referrer      string `json:"referrer"`
	AffiliateCode string `json:"affiliate_code"`
	ClientIP      string `json:"client_ip"`
	DeviceType    string `json:"device_type"`
	CreatedAt     string `json:"created_at"`
}

// Record 记录一次访问，失败只影响统计，不影响业务主流程。
func (s *SiteVisitService) Record(input RecordSiteVisitInput) error {
	if s == nil || s.repo == nil {
		return nil
	}
	path := cleanVisitPath(input.Path)
	if path == "" {
		path = "/"
	}
	sourceType := normalizeVisitSource(input.SourceType, input.AffiliateCode, input.Referrer)
	deviceType := normalizeVisitDevice(input.DeviceType, input.UserAgent)
	pageType := normalizeVisitPageType(input.PageType, path)

	return s.repo.Create(&models.SiteVisitEvent{
		VisitorKey:    truncateString(strings.TrimSpace(input.VisitorKey), 80),
		SessionKey:    truncateString(strings.TrimSpace(input.SessionKey), 80),
		UserID:        input.UserID,
		Path:          truncateString(path, 512),
		PageType:      pageType,
		SourceType:    sourceType,
		Referrer:      truncateString(strings.TrimSpace(input.Referrer), 2000),
		AffiliateCode: truncateString(strings.ToUpper(strings.TrimSpace(input.AffiliateCode)), 64),
		ClientIP:      truncateString(strings.TrimSpace(input.ClientIP), 80),
		UserAgent:     truncateString(strings.TrimSpace(input.UserAgent), 2000),
		DeviceType:    deviceType,
		CreatedAt:     time.Now(),
	})
}

// GetSummary 获取汇总。
func (s *SiteVisitService) GetSummary(input SiteVisitQueryInput) (*SiteVisitSummaryResponse, error) {
	if s == nil || s.repo == nil {
		return &SiteVisitSummaryResponse{}, nil
	}
	window, err := resolveSiteVisitWindow(input, time.Now())
	if err != nil {
		return nil, err
	}
	row, err := s.repo.GetSummary(window.filter)
	if err != nil {
		return nil, err
	}
	return &SiteVisitSummaryResponse{
		Range:    window.rangeKey,
		From:     window.startAt.Format(time.RFC3339),
		To:       window.endAt.Add(-time.Second).Format(time.RFC3339),
		Timezone: window.timezone,
		KPI: SiteVisitSummaryKPI{
			PV:            row.PV,
			UV:            row.UV,
			UniqueIP:      row.UniqueIP,
			AffiliatePV:   row.AffiliatePV,
			LoggedInPV:    row.LoggedInPV,
			VisitorPV:     row.VisitorPV,
			MobilePV:      row.MobilePV,
			DesktopPV:     row.DesktopPV,
			TabletPV:      row.TabletPV,
			DirectPV:      row.DirectPV,
			SearchPV:      row.SearchPV,
			ExternalPV:    row.ExternalPV,
			UnknownPV:     row.UnknownPV,
			AffiliateRate: formatRate(row.AffiliatePV, row.PV),
			LoggedInRate:  formatRate(row.LoggedInPV, row.PV),
			MobileRate:    formatRate(row.MobilePV, row.PV),
		},
	}, nil
}

// GetTrends 获取趋势。
func (s *SiteVisitService) GetTrends(input SiteVisitQueryInput) (*SiteVisitTrendResponse, error) {
	if s == nil || s.repo == nil {
		return &SiteVisitTrendResponse{}, nil
	}
	window, err := resolveSiteVisitWindow(input, time.Now())
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetTrends(window.filter)
	if err != nil {
		return nil, err
	}
	points := fillSiteVisitTrendPoints(window, rows)
	return &SiteVisitTrendResponse{
		Range:    window.rangeKey,
		From:     window.startAt.Format(time.RFC3339),
		To:       window.endAt.Add(-time.Second).Format(time.RFC3339),
		Timezone: window.timezone,
		Points:   points,
	}, nil
}

// GetSources 获取来源统计。
func (s *SiteVisitService) GetSources(input SiteVisitQueryInput) ([]SiteVisitSourceItem, error) {
	if s == nil || s.repo == nil {
		return []SiteVisitSourceItem{}, nil
	}
	window, err := resolveSiteVisitWindow(input, time.Now())
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetSources(window.filter)
	if err != nil {
		return nil, err
	}
	var total int64
	for _, row := range rows {
		total += row.PV
	}
	items := make([]SiteVisitSourceItem, 0, len(rows))
	for _, row := range rows {
		source := normalizeVisitSource(row.SourceType, "", "")
		items = append(items, SiteVisitSourceItem{
			SourceType: source,
			PV:         row.PV,
			UV:         row.UV,
			Rate:       formatRate(row.PV, total),
		})
	}
	return items, nil
}

// GetPages 获取页面排行。
func (s *SiteVisitService) GetPages(input SiteVisitQueryInput, limit int) ([]SiteVisitPageItem, error) {
	if s == nil || s.repo == nil {
		return []SiteVisitPageItem{}, nil
	}
	window, err := resolveSiteVisitWindow(input, time.Now())
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetPages(window.filter, limit)
	if err != nil {
		return nil, err
	}
	var total int64
	for _, row := range rows {
		total += row.PV
	}
	items := make([]SiteVisitPageItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, SiteVisitPageItem{
			Path: row.Path,
			PV:   row.PV,
			UV:   row.UV,
			Rate: formatRate(row.PV, total),
		})
	}
	return items, nil
}

// ListRecent 获取最近访问。
func (s *SiteVisitService) ListRecent(input SiteVisitQueryInput) ([]SiteVisitRecentItem, int64, error) {
	if s == nil || s.repo == nil {
		return []SiteVisitRecentItem{}, 0, nil
	}
	window, err := resolveSiteVisitWindow(input, time.Now())
	if err != nil {
		return nil, 0, err
	}
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	rows, total, err := s.repo.ListRecent(window.filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]SiteVisitRecentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, SiteVisitRecentItem{
			ID:            row.ID,
			VisitorKey:    row.VisitorKey,
			SessionKey:    row.SessionKey,
			UserID:        row.UserID,
			Path:          row.Path,
			PageType:      row.PageType,
			SourceType:    row.SourceType,
			Referrer:      row.Referrer,
			AffiliateCode: row.AffiliateCode,
			ClientIP:      row.ClientIP,
			DeviceType:    row.DeviceType,
			CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, total, nil
}

type siteVisitWindow struct {
	rangeKey string
	startAt  time.Time
	endAt    time.Time
	timezone string
	filter   repository.SiteVisitFilter
}

func resolveSiteVisitWindow(input SiteVisitQueryInput, now time.Time) (siteVisitWindow, error) {
	tz := strings.TrimSpace(input.Timezone)
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
		tz = "UTC"
	}

	localNow := now.In(loc)
	rangeKey := strings.ToLower(strings.TrimSpace(input.Range))
	if rangeKey == "" {
		rangeKey = "7d"
	}

	var startLocal, endLocal time.Time
	switch rangeKey {
	case "today":
		startLocal = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
		endLocal = startLocal.AddDate(0, 0, 1)
	case "week", "this_week":
		weekday := int(localNow.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
		startLocal = today.AddDate(0, 0, -(weekday - 1))
		endLocal = startLocal.AddDate(0, 0, 7)
	case "month", "this_month":
		startLocal = time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, loc)
		endLocal = startLocal.AddDate(0, 1, 0)
	case "last_month":
		thisMonth := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, loc)
		startLocal = thisMonth.AddDate(0, -1, 0)
		endLocal = thisMonth
	case "custom":
		if input.From == nil || input.To == nil {
			return siteVisitWindow{}, ErrSiteVisitRangeInvalid
		}
		startLocal = input.From.In(loc)
		endLocal = input.To.In(loc)
		if !hasClock(input.To.In(loc)) {
			endLocal = endLocal.AddDate(0, 0, 1)
		}
	default:
		startLocal = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6)
		endLocal = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		rangeKey = "7d"
	}

	if !startLocal.Before(endLocal) {
		return siteVisitWindow{}, ErrSiteVisitRangeInvalid
	}
	if endLocal.Sub(startLocal) > siteVisitCustomMaxDays*24*time.Hour {
		return siteVisitWindow{}, ErrSiteVisitRangeInvalid
	}

	startAt := startLocal.UTC()
	endAt := endLocal.UTC()
	return siteVisitWindow{
		rangeKey: rangeKey,
		startAt:  startAt,
		endAt:    endAt,
		timezone: tz,
		filter: repository.SiteVisitFilter{
			StartAt:    startAt,
			EndAt:      endAt,
			Timezone:   tz,
			SourceType: normalizeOptionalFilter(input.SourceType),
			DeviceType: normalizeOptionalFilter(input.DeviceType),
			PageType:   normalizeOptionalFilter(input.PageType),
		},
	}, nil
}

func fillSiteVisitTrendPoints(window siteVisitWindow, rows []repository.SiteVisitTrendRow) []SiteVisitTrendPoint {
	loc, err := time.LoadLocation(window.timezone)
	if err != nil {
		loc = time.UTC
	}
	rowByDay := make(map[string]repository.SiteVisitTrendRow, len(rows))
	for _, row := range rows {
		rowByDay[row.Day] = row
	}
	start := window.startAt.In(loc)
	end := window.endAt.In(loc)
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)
	points := make([]SiteVisitTrendPoint, 0, int(endDay.Sub(startDay).Hours()/24)+1)
	for day := startDay; day.Before(endDay); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		row := rowByDay[key]
		points = append(points, SiteVisitTrendPoint{
			Date:        key,
			PV:          row.PV,
			UV:          row.UV,
			AffiliatePV: row.AffiliatePV,
			DirectPV:    row.DirectPV,
			SearchPV:    row.SearchPV,
			ExternalPV:  row.ExternalPV,
			UnknownPV:   row.UnknownPV,
		})
	}
	return points
}

func normalizeVisitSource(raw, affiliateCode, referrer string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case SiteVisitSourceDirect, SiteVisitSourceAffiliate, SiteVisitSourceSearch, SiteVisitSourceExternal, SiteVisitSourceUnknown:
		return value
	}
	if strings.TrimSpace(affiliateCode) != "" {
		return SiteVisitSourceAffiliate
	}
	trimmedReferrer := strings.TrimSpace(referrer)
	if trimmedReferrer == "" {
		return SiteVisitSourceDirect
	}
	host := ""
	if parsed, err := url.Parse(trimmedReferrer); err == nil {
		host = strings.ToLower(parsed.Hostname())
	}
	if host == "" {
		return SiteVisitSourceUnknown
	}
	if strings.Contains(host, "google.") || strings.Contains(host, "bing.") || strings.Contains(host, "baidu.") ||
		strings.Contains(host, "yahoo.") || strings.Contains(host, "duckduckgo.") || strings.Contains(host, "sogou.") ||
		strings.Contains(host, "so.com") || strings.Contains(host, "yandex.") {
		return SiteVisitSourceSearch
	}
	return SiteVisitSourceExternal
}

func normalizeVisitDevice(raw, userAgent string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case SiteVisitDeviceDesktop, SiteVisitDeviceMobile, SiteVisitDeviceTablet, SiteVisitDeviceUnknown:
		return value
	}
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet"):
		return SiteVisitDeviceTablet
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || strings.Contains(ua, "android"):
		return SiteVisitDeviceMobile
	case ua != "":
		return SiteVisitDeviceDesktop
	default:
		return SiteVisitDeviceUnknown
	}
}

func normalizeVisitPageType(raw, path string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value != "" && len(value) <= 40 {
		return value
	}
	if path == "/" {
		return "home"
	}
	if strings.HasPrefix(path, "/products/") {
		return "product"
	}
	if strings.HasPrefix(path, "/categories/") {
		return "category"
	}
	if strings.HasPrefix(path, "/cart") {
		return "cart"
	}
	if strings.HasPrefix(path, "/checkout") || strings.HasPrefix(path, "/pay") {
		return "checkout"
	}
	if strings.HasPrefix(path, "/me") {
		return "account"
	}
	return "page"
}

func cleanVisitPath(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		if parsed, err := url.Parse(value); err == nil {
			value = parsed.RequestURI()
		}
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return value
}

func normalizeOptionalFilter(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || value == "all" || value == "__all__" {
		return ""
	}
	return value
}

func truncateString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}

func hasClock(t time.Time) bool {
	return t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 || t.Nanosecond() != 0
}

func formatRate(part, total int64) string {
	if total <= 0 {
		return "0.00"
	}
	return formatPercentValue(float64(part) / float64(total) * 100)
}
