package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateInviteCode 创建平台代理邀请码
func (r *GormAffiliateRepository) CreateInviteCode(code *models.AffiliateInviteCode) error {
	return r.db.Create(code).Error
}

// GetInviteCodeByCode 按邀请码查询平台代理邀请码
func (r *GormAffiliateRepository) GetInviteCodeByCode(code string) (*models.AffiliateInviteCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return nil, nil
	}
	var row models.AffiliateInviteCode
	if err := r.db.Preload("UsedByUser").Where("code = ?", normalized).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetInviteCodeByCodeForUpdate 按邀请码查询并锁定平台代理邀请码
func (r *GormAffiliateRepository) GetInviteCodeByCodeForUpdate(code string) (*models.AffiliateInviteCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return nil, nil
	}
	var row models.AffiliateInviteCode
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", normalized).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetInviteCodeByID 按ID查询平台代理邀请码
func (r *GormAffiliateRepository) GetInviteCodeByID(id uint) (*models.AffiliateInviteCode, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.AffiliateInviteCode
	if err := r.db.Preload("UsedByUser").First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListInviteCodes 查询平台代理邀请码列表
func (r *GormAffiliateRepository) ListInviteCodes(filter AffiliateInviteCodeListFilter) ([]models.AffiliateInviteCode, int64, error) {
	query := r.db.Model(&models.AffiliateInviteCode{}).Preload("UsedByUser")
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(code LIKE ? OR remark LIKE ?)", strings.ToUpper(like), like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = applyPagination(query, filter.Page, filter.PageSize)

	var rows []models.AffiliateInviteCode
	if err := query.Order("id desc").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateInviteCodeStatus 更新平台代理邀请码状态
func (r *GormAffiliateRepository) UpdateInviteCodeStatus(id uint, status string, updatedAt time.Time) error {
	if id == 0 {
		return nil
	}
	return r.db.Model(&models.AffiliateInviteCode{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     strings.TrimSpace(status),
			"updated_at": updatedAt,
		}).Error
}

// MarkInviteCodeUsed 标记平台代理邀请码已使用
func (r *GormAffiliateRepository) MarkInviteCodeUsed(id, userID uint, usedAt time.Time) (int64, error) {
	if id == 0 || userID == 0 {
		return 0, nil
	}
	result := r.db.Model(&models.AffiliateInviteCode{}).
		Where("id = ? AND status = ?", id, constants.AffiliateInviteCodeStatusActive).
		Updates(map[string]interface{}{
			"status":          constants.AffiliateInviteCodeStatusUsed,
			"used_by_user_id": userID,
			"used_at":         usedAt,
			"updated_at":      usedAt,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
