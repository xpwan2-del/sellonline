package public

import (
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// SiteVisitTrackRequest 站点访问上报请求。
type SiteVisitTrackRequest struct {
	VisitorKey    string `json:"visitor_key"`
	SessionKey    string `json:"session_key"`
	UserID        uint   `json:"user_id"`
	Path          string `json:"path"`
	PageType      string `json:"page_type"`
	SourceType    string `json:"source_type"`
	Referrer      string `json:"referrer"`
	AffiliateCode string `json:"affiliate_code"`
	DeviceType    string `json:"device_type"`
}

// TrackSiteVisit 记录站点访问。
func (h *Handler) TrackSiteVisit(c *gin.Context) {
	var req SiteVisitTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	if h.SiteVisitService != nil {
		if err := h.SiteVisitService.Record(service.RecordSiteVisitInput{
			VisitorKey:    req.VisitorKey,
			SessionKey:    req.SessionKey,
			UserID:        req.UserID,
			Path:          req.Path,
			PageType:      req.PageType,
			SourceType:    req.SourceType,
			Referrer:      req.Referrer,
			AffiliateCode: strings.TrimSpace(req.AffiliateCode),
			ClientIP:      c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			DeviceType:    req.DeviceType,
		}); err != nil {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
			return
		}
	}
	response.Success(c, gin.H{"ok": true})
}
