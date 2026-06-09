package models

import "time"

// SiteVisitEvent 站点访问事件。
// 说明：这是独立统计表，不设置外键，避免影响订单、用户、返佣等核心业务。
type SiteVisitEvent struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	VisitorKey    string    `gorm:"index;size:80;default:''" json:"visitor_key"`
	SessionKey    string    `gorm:"index;size:80;default:''" json:"session_key"`
	UserID        uint      `gorm:"index;not null;default:0" json:"user_id"`
	Path          string    `gorm:"size:512;not null;default:''" json:"path"`
	PageType      string    `gorm:"index;size:40;not null;default:'page'" json:"page_type"`
	SourceType    string    `gorm:"index;size:40;not null;default:'unknown'" json:"source_type"`
	Referrer      string    `gorm:"type:text" json:"referrer"`
	AffiliateCode string    `gorm:"index;size:64;not null;default:''" json:"affiliate_code"`
	ClientIP      string    `gorm:"index;size:80;not null;default:''" json:"client_ip"`
	UserAgent     string    `gorm:"type:text" json:"user_agent"`
	DeviceType    string    `gorm:"index;size:40;not null;default:'unknown'" json:"device_type"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名。
func (SiteVisitEvent) TableName() string {
	return "site_visit_events"
}
