package public

import (
	"strings"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// ListMyAffiliateCustomers 查询当前一级代理商名下客户。
func (h *Handler) ListMyAffiliateCustomers(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	registeredFrom, registeredTo, err := shared.ParseQueryTimeRange(c, "registered_from", "registered_to")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateService.ListMyCustomers(uid, service.AffiliateCustomerRelationListInput{
		Page:           page,
		PageSize:       pageSize,
		Keyword:        strings.TrimSpace(c.Query("keyword")),
		RegisteredFrom: registeredFrom,
		RegisteredTo:   registeredTo,
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, affiliateCustomerRespList(rows), response.BuildPagination(page, pageSize, total))
}

func affiliateCustomerRespList(rows []service.AffiliateCustomerSummary) []dto.AffiliateCustomerResp {
	result := make([]dto.AffiliateCustomerResp, 0, len(rows))
	for i := range rows {
		result = append(result, dto.AffiliateCustomerResp{
			ID:               rows[i].ID,
			Email:            rows[i].Email,
			DisplayName:      rows[i].DisplayName,
			RegisteredAt:     rows[i].RegisteredAt,
			LastOrderAt:      rows[i].LastOrderAt,
			OrderCount:       rows[i].OrderCount,
			CommissionAmount: rows[i].CommissionAmount,
		})
	}
	return result
}
