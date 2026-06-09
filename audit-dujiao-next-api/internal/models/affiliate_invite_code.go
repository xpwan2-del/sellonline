package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateInviteCode 平台一级代理邀请码
type AffiliateInviteCode struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Code             string         `gorm:"type:varchar(32);not null;uniqueIndex" json:"code"`
	Status           string         `gorm:"type:varchar(20);not null;index" json:"status"`
	UsedByUserID     *uint          `gorm:"index" json:"used_by_user_id,omitempty"`
	UsedAt           *time.Time     `json:"used_at,omitempty"`
	CreatedByAdminID uint           `gorm:"index;not null;default:0" json:"created_by_admin_id"`
	Remark           string         `gorm:"type:varchar(255);not null;default:''" json:"remark"`
	CreatedAt        time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	UsedByUser User `gorm:"foreignKey:UsedByUserID" json:"used_by_user,omitempty"`
}

// TableName 指定表名
func (AffiliateInviteCode) TableName() string {
	return "affiliate_invite_codes"
}
