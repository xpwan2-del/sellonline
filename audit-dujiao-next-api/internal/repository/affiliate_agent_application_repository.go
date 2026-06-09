package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// CreateAgentApplication 创建平台代理申请。
func (r *GormAffiliateRepository) CreateAgentApplication(app *models.AffiliateAgentApplication) error {
	return r.db.Create(app).Error
}

// GetLatestAgentApplicationByUserID 获取用户最新平台代理申请。
func (r *GormAffiliateRepository) GetLatestAgentApplicationByUserID(userID uint) (*models.AffiliateAgentApplication, error) {
	if userID == 0 {
		return nil, nil
	}
	var row models.AffiliateAgentApplication
	if err := r.db.Preload("User").Where("user_id = ?", userID).Order("id desc").First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetAgentApplicationByID 按 ID 获取平台代理申请。
func (r *GormAffiliateRepository) GetAgentApplicationByID(id uint) (*models.AffiliateAgentApplication, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.AffiliateAgentApplication
	if err := r.db.Preload("User").First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListAgentApplications 查询平台代理申请列表。
func (r *GormAffiliateRepository) ListAgentApplications(filter AffiliateAgentApplicationListFilter) ([]models.AffiliateAgentApplication, int64, error) {
	query := r.db.Model(&models.AffiliateAgentApplication{}).Preload("User")
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(email LIKE ? OR phone LIKE ? OR message LIKE ?)", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = applyPagination(query, filter.Page, filter.PageSize)

	var rows []models.AffiliateAgentApplication
	if err := query.Order("id desc").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateAgentApplicationStatus 更新平台代理申请状态和后台备注。
func (r *GormAffiliateRepository) UpdateAgentApplicationStatus(id uint, status, adminNote string, updatedAt time.Time) error {
	if id == 0 {
		return nil
	}
	return r.db.Model(&models.AffiliateAgentApplication{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     strings.TrimSpace(status),
			"admin_note": strings.TrimSpace(adminNote),
			"updated_at": updatedAt,
		}).Error
}
