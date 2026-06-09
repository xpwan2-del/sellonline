package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateCommissionItem 推广返利商品行佣金明细
type AffiliateCommissionItem struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	AffiliateCommissionID uint           `gorm:"not null;index;uniqueIndex:idx_affiliate_commission_item_unique" json:"affiliate_commission_id"`
	OrderItemID           uint           `gorm:"not null;index;uniqueIndex:idx_affiliate_commission_item_unique" json:"order_item_id"`
	ProductID             uint           `gorm:"not null;index" json:"product_id"`
	BaseAmount            Money          `gorm:"type:decimal(20,2);not null;default:0" json:"base_amount"`
	RatePercent           Money          `gorm:"type:decimal(10,2);not null;default:0" json:"rate_percent"`
	CommissionAmount      Money          `gorm:"type:decimal(20,2);not null;default:0" json:"commission_amount"`
	CreatedAt             time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	AffiliateCommission AffiliateCommission `gorm:"foreignKey:AffiliateCommissionID" json:"-"`
	OrderItem           OrderItem           `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	Product             Product             `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 指定表名
func (AffiliateCommissionItem) TableName() string {
	return "affiliate_commission_items"
}
