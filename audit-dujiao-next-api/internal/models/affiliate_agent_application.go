package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateAgentApplication 记录普通用户提交的平台代理申请。
type AffiliateAgentApplication struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Email     string         `gorm:"type:varchar(255);not null;index" json:"email"`
	Phone     string         `gorm:"type:varchar(64);not null;default:''" json:"phone"`
	Message   string         `gorm:"type:text;not null;default:''" json:"message"`
	Status    string         `gorm:"type:varchar(20);not null;index" json:"status"`
	AdminNote string         `gorm:"type:varchar(255);not null;default:''" json:"admin_note"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (AffiliateAgentApplication) TableName() string {
	return "affiliate_agent_applications"
}
