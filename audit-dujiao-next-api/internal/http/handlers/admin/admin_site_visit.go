package admin

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// GetSiteVisitSummary 获取进站统计汇总。
func (h *Handler) GetSiteVisitSummary(c *gin.Context) {
	input, err := parseSiteVisitQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.SiteVisitService.GetSummary(input)
	if err != nil {
		respondSiteVisitError(c, err)
		return
	}
	response.Success(c, data)
}

// GetSiteVisitTrends 获取进站趋势。
func (h *Handler) GetSiteVisitTrends(c *gin.Context) {
	input, err := parseSiteVisitQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.SiteVisitService.GetTrends(input)
	if err != nil {
		respondSiteVisitError(c, err)
		return
	}
	response.Success(c, data)
}

// GetSiteVisitSources 获取来源统计。
func (h *Handler) GetSiteVisitSources(c *gin.Context) {
	input, err := parseSiteVisitQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.SiteVisitService.GetSources(input)
	if err != nil {
		respondSiteVisitError(c, err)
		return
	}
	response.Success(c, data)
}

// GetSiteVisitPages 获取页面排行。
func (h *Handler) GetSiteVisitPages(c *gin.Context) {
	input, err := parseSiteVisitQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.SiteVisitService.GetPages(input, 10)
	if err != nil {
		respondSiteVisitError(c, err)
		return
	}
	response.Success(c, data)
}

// GetSiteVisitRecent 获取最近访问列表。
func (h *Handler) GetSiteVisitRecent(c *gin.Context) {
	input, err := parseSiteVisitQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	input.Page, input.PageSize = shared.ParsePagination(c)
	rows, total, err := h.SiteVisitService.ListRecent(input)
	if err != nil {
		respondSiteVisitError(c, err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(input.Page, input.PageSize, total))
}

func respondSiteVisitError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrSiteVisitRangeInvalid) {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	shared.RespondError(c, response.CodeInternal, "error.dashboard_fetch_failed", err)
}

func parseSiteVisitQuery(c *gin.Context) (service.SiteVisitQueryInput, error) {
	rangeRaw := strings.TrimSpace(c.DefaultQuery("range", "7d"))
	timezone := strings.TrimSpace(c.Query("tz"))
	from, err := parseSiteVisitTime(strings.TrimSpace(c.Query("start_date")))
	if err != nil {
		return service.SiteVisitQueryInput{}, err
	}
	to, err := parseSiteVisitTime(strings.TrimSpace(c.Query("end_date")))
	if err != nil {
		return service.SiteVisitQueryInput{}, err
	}
	return service.SiteVisitQueryInput{
		Range:      rangeRaw,
		From:       from,
		To:         to,
		Timezone:   timezone,
		SourceType: strings.TrimSpace(c.Query("source_type")),
		DeviceType: strings.TrimSpace(c.Query("device_type")),
		PageType:   strings.TrimSpace(c.Query("page_type")),
	}, nil
}

func parseSiteVisitTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return &parsed, nil
	}
	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		return &parsed, nil
	}
	return nil, service.ErrSiteVisitRangeInvalid
}
