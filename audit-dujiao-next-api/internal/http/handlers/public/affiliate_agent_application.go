package public

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

type affiliateAgentApplicationRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Message string `json:"message"`
}

// GetMyAffiliateAgentApplication 获取当前用户最新平台代理申请。
func (h *Handler) GetMyAffiliateAgentApplication(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	row, err := h.AffiliateService.GetLatestAgentApplicationByUserID(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, affiliateAgentApplicationResp(row))
}

// CreateAffiliateAgentApplication 提交平台代理申请。
func (h *Handler) CreateAffiliateAgentApplication(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	var req affiliateAgentApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.AffiliateService.CreateAgentApplication(service.AffiliateAgentApplicationCreateInput{
		UserID:  uid,
		Phone:   req.Phone,
		Message: req.Message,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAffiliateAgentApplicationInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		case errors.Is(err, service.ErrAffiliateAgentApplicationExists):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, affiliateAgentApplicationResp(row))
}

func affiliateAgentApplicationResp(row *models.AffiliateAgentApplication) gin.H {
	if row == nil || row.ID == 0 {
		return gin.H{}
	}
	return gin.H{
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
}
