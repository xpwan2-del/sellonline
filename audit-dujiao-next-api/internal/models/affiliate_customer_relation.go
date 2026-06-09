package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateCustomerRelation 记录买家长期归属的一级代理商
type AffiliateCustomerRelation struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CustomerUserID      uint           `gorm:"not null;uniqueIndex" json:"customer_user_id"`
	AffiliateProfileID  uint           `gorm:"not null;index" json:"affiliate_profile_id"`
	SourceAffiliateCode string         `gorm:"type:varchar(32);not null;default:''" json:"source_affiliate_code"`
	SourceOrderID       *uint          `gorm:"index" json:"source_order_id,omitempty"`
	SourceType          string         `gorm:"type:varchar(20);not null;index" json:"source_type"`
	BoundAt             time.Time      `gorm:"index;not null" json:"bound_at"`
	CreatedAt           time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Customer         User             `gorm:"foreignKey:CustomerUserID" json:"customer,omitempty"`
	AffiliateProfile AffiliateProfile `gorm:"foreignKey:AffiliateProfileID" json:"affiliate_profile,omitempty"`
}

// TableName 指定表名
func (AffiliateCustomerRelation) TableName() string {
	return "affiliate_customer_relations"
}
