package public

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// GetMyAffiliateReportSummary 获取我的推广返利统计报表概览。
func (h *Handler) GetMyAffiliateReportSummary(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.AffiliateReportService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	query, err := parseMyAffiliateReportQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.AffiliateReportService.GetUserSummary(uid, query, requestOrigin(c))
	if err != nil {
		if errors.Is(err, service.ErrAffiliateReportRangeInvalid) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, data)
}

// ListMyAffiliateReportCommissions 获取我的推广返利统计报表明细。
func (h *Handler) ListMyAffiliateReportCommissions(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.AffiliateReportService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	query, err := parseMyAffiliateReportQuery(c)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateReportService.ListUserCommissions(uid, service.AffiliateReportCommissionQuery{
		AffiliateReportQuery: query,
		Page:                 page,
		PageSize:             pageSize,
	})
	if err != nil {
		if errors.Is(err, service.ErrAffiliateReportRangeInvalid) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

func parseMyAffiliateReportQuery(c *gin.Context) (service.AffiliateReportQuery, error) {
	startAt, endAt, err := shared.ParseQueryTimeRange(c, "start_date", "end_date")
	if err != nil {
		return service.AffiliateReportQuery{}, err
	}
	productID, err := shared.ParseQueryUint(c.Query("product_id"), false)
	if err != nil {
		return service.AffiliateReportQuery{}, err
	}
	return service.AffiliateReportQuery{
		StartAt:   startAt,
		EndAt:     endAt,
		ProductID: productID,
		Status:    strings.TrimSpace(c.Query("status")),
	}, nil
}

func requestOrigin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	scheme := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = strings.TrimSpace(c.GetHeader("X-Scheme"))
	}
	if scheme == "" {
		scheme = "https"
		if c.Request != nil && c.Request.TLS == nil {
			scheme = "http"
		}
	}
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" && c.Request != nil {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}
