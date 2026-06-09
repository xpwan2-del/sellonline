package admin

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

type createAffiliateInviteCodeRequest struct {
	Code   string `json:"code"`
	Remark string `json:"remark"`
}

type updateAffiliateInviteCodeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ListAffiliateInviteCodes 查询平台代理邀请码
func (h *Handler) ListAffiliateInviteCodes(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.AffiliateService.ListInviteCodes(service.AffiliateInviteCodeListInput{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, affiliateInviteCodeRespList(rows), response.BuildPagination(page, pageSize, total))
}

// CreateAffiliateInviteCode 创建平台代理邀请码
func (h *Handler) CreateAffiliateInviteCode(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	var req createAffiliateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.AffiliateService.CreateInviteCode(service.AffiliateInviteCodeCreateInput{
		Code:             req.Code,
		Remark:           req.Remark,
		CreatedByAdminID: adminID,
	})
	if err != nil {
		if errors.Is(err, service.ErrAffiliateCodeInvalid) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, affiliateInviteCodeResp(row))
}

// UpdateAffiliateInviteCodeStatus 更新平台代理邀请码状态
func (h *Handler) UpdateAffiliateInviteCodeStatus(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	var req updateAffiliateInviteCodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.AffiliateService.UpdateInviteCodeStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
		case errors.Is(err, service.ErrAffiliateInviteCodeInvalid),
			errors.Is(err, service.ErrAffiliateInviteCodeUnavailable):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, affiliateInviteCodeResp(row))
}

// GetAffiliateInviteCodeUsage 查询平台代理邀请码使用情况
func (h *Handler) GetAffiliateInviteCodeUsage(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.AffiliateRepo.GetInviteCodeByID(id)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	if row == nil {
		shared.RespondError(c, response.CodeNotFound, "error.not_found", nil)
		return
	}
	response.Success(c, affiliateInviteCodeResp(row))
}

func affiliateInviteCodeRespList(rows []models.AffiliateInviteCode) []gin.H {
	items := make([]gin.H, 0, len(rows))
	for i := range rows {
		items = append(items, affiliateInviteCodeResp(&rows[i]))
	}
	return items
}

func affiliateInviteCodeResp(row *models.AffiliateInviteCode) gin.H {
	if row == nil {
		return gin.H{}
	}
	usable := strings.TrimSpace(row.Status) == constants.AffiliateInviteCodeStatusActive && row.UsedByUserID == nil && row.UsedAt == nil
	data := gin.H{
		"id":                  row.ID,
		"code":                row.Code,
		"status":              row.Status,
		"usable":              usable,
		"used_by_user_id":     row.UsedByUserID,
		"used_at":             row.UsedAt,
		"created_by_admin_id": row.CreatedByAdminID,
		"remark":              row.Remark,
		"created_at":          row.CreatedAt,
		"updated_at":          row.UpdatedAt,
	}
	if row.UsedByUser.ID > 0 {
		data["used_by_user"] = gin.H{
			"id":           row.UsedByUser.ID,
			"email":        row.UsedByUser.Email,
			"display_name": row.UsedByUser.DisplayName,
		}
	}
	return data
}
