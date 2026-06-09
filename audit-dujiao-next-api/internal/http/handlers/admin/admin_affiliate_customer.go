package admin

import (
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// ListAffiliateCustomers 查询一级代理商名下买家
func (h *Handler) ListAffiliateCustomers(c *gin.Context) {
	affiliateProfileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateService.ListCustomerRelations(service.AffiliateCustomerRelationListInput{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: affiliateProfileID,
		Keyword:            strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, affiliateCustomerRelationRespList(rows), response.BuildPagination(page, pageSize, total))
}

func affiliateCustomerRelationRespList(rows []models.AffiliateCustomerRelation) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for i := range rows {
		row := rows[i]
		items = append(items, gin.H{
			"id":                    row.ID,
			"customer_user_id":      row.CustomerUserID,
			"affiliate_profile_id":  row.AffiliateProfileID,
			"source_affiliate_code": row.SourceAffiliateCode,
			"source_order_id":       row.SourceOrderID,
			"source_type":           row.SourceType,
			"bound_at":              row.BoundAt,
			"created_at":            row.CreatedAt,
			"customer": gin.H{
				"id":           row.Customer.ID,
				"email":        row.Customer.Email,
				"display_name": row.Customer.DisplayName,
				"status":       row.Customer.Status,
			},
		})
	}
	return items
}
