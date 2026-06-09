package admin

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// GetAffiliateReportSummary 获取推广返利统计报表概览
func (h *Handler) GetAffiliateReportSummary(c *gin.Context) {
	if h.AffiliateReportService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	query, err := parseAffiliateReportQuery(c, true)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	data, err := h.AffiliateReportService.GetAdminSummary(query)
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

// ListAffiliateReportCommissions 获取推广返利统计报表明细
func (h *Handler) ListAffiliateReportCommissions(c *gin.Context) {
	if h.AffiliateReportService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	query, err := parseAffiliateReportQuery(c, true)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateReportService.ListAdminCommissions(service.AffiliateReportCommissionQuery{
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

func parseAffiliateReportQuery(c *gin.Context, allowProfile bool) (service.AffiliateReportQuery, error) {
	startAt, endAt, err := shared.ParseQueryTimeRange(c, "start_date", "end_date")
	if err != nil {
		return service.AffiliateReportQuery{}, err
	}
	productID, err := shared.ParseQueryUint(c.Query("product_id"), false)
	if err != nil {
		return service.AffiliateReportQuery{}, err
	}
	var profileID uint
	if allowProfile {
		profileID, err = shared.ParseQueryUint(c.Query("affiliate_profile_id"), false)
		if err != nil {
			return service.AffiliateReportQuery{}, err
		}
	}
	return service.AffiliateReportQuery{
		StartAt:            startAt,
		EndAt:              endAt,
		AffiliateProfileID: profileID,
		AffiliateKeyword:   strings.TrimSpace(c.Query("affiliate_keyword")),
		ProductID:          productID,
		Status:             strings.TrimSpace(c.Query("status")),
	}, nil
}
