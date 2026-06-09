package admin

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

type updateAffiliateAgentApplicationStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	AdminNote string `json:"admin_note"`
}

// ListAffiliateAgentApplications 查询平台代理申请列表。
func (h *Handler) ListAffiliateAgentApplications(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateService.ListAgentApplications(service.AffiliateAgentApplicationListInput{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, affiliateAgentApplicationRespList(rows), response.BuildPagination(page, pageSize, total))
}

// UpdateAffiliateAgentApplicationStatus 更新平台代理申请状态。
func (h *Handler) UpdateAffiliateAgentApplicationStatus(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	var req updateAffiliateAgentApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.AffiliateService.UpdateAgentApplicationStatus(id, service.AffiliateAgentApplicationStatusInput{
		Status:    req.Status,
		AdminNote: req.AdminNote,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
		case errors.Is(err, service.ErrAffiliateAgentApplicationInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, affiliateAgentApplicationAdminResp(row))
}

func affiliateAgentApplicationRespList(rows []models.AffiliateAgentApplication) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for i := range rows {
		items = append(items, affiliateAgentApplicationAdminResp(&rows[i]))
	}
	return items
}

func affiliateAgentApplicationAdminResp(row *models.AffiliateAgentApplication) gin.H {
	if row == nil || row.ID == 0 {
		return gin.H{}
	}
	data := gin.H{
		"id":         row.ID,
		"user_id":    row.UserID,
		"email":      row.Email,
		"phone":      row.Phone,
		"message":    row.Message,
		"status":     strings.TrimSpace(row.Status),
		"admin_note": row.AdminNote,
		"created_at": row.CreatedAt,
		"updated_at": row.UpdatedAt,
	}
	if row.User.ID > 0 {
		data["user"] = gin.H{
			"id":           row.User.ID,
			"email":        row.User.Email,
			"display_name": row.User.DisplayName,
		}
	}
	return data
}
