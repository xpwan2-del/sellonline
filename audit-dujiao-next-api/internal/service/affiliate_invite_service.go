package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

// AffiliateInviteCodeListInput 平台代理邀请码列表输入
type AffiliateInviteCodeListInput struct {
	Page     int
	PageSize int
	Status   string
	Keyword  string
}

// AffiliateInviteCodeCreateInput 创建平台代理邀请码输入
type AffiliateInviteCodeCreateInput struct {
	Code             string
	Remark           string
	CreatedByAdminID uint
}

// ListInviteCodes 查询平台代理邀请码列表
func (s *AffiliateService) ListInviteCodes(input AffiliateInviteCodeListInput) ([]models.AffiliateInviteCode, int64, error) {
	if s.repo == nil {
		return nil, 0, ErrNotFound
	}
	return s.repo.ListInviteCodes(repository.AffiliateInviteCodeListFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Status:   strings.TrimSpace(input.Status),
		Keyword:  strings.TrimSpace(input.Keyword),
	})
}

// CreateInviteCode 创建平台代理邀请码
func (s *AffiliateService) CreateInviteCode(input AffiliateInviteCodeCreateInput) (*models.AffiliateInviteCode, error) {
	if s.repo == nil {
		return nil, ErrNotFound
	}
	now := time.Now()
	rawInputCode := strings.TrimSpace(input.Code)
	inputCode := normalizeAffiliateInviteCode(input.Code)
	if rawInputCode != "" && inputCode == "" {
		return nil, ErrAffiliateCodeInvalid
	}
	if inputCode != "" {
		row := &models.AffiliateInviteCode{
			Code:             inputCode,
			Status:           constants.AffiliateInviteCodeStatusActive,
			CreatedByAdminID: input.CreatedByAdminID,
			Remark:           strings.TrimSpace(input.Remark),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.repo.CreateInviteCode(row); err != nil {
			if isUniqueViolation(err) {
				return nil, ErrAffiliateCodeInvalid
			}
			return nil, err
		}
		return s.repo.GetInviteCodeByID(row.ID)
	}

	const maxRetry = 8
	for i := 0; i < maxRetry; i++ {
		code, err := generateAffiliateCode()
		if err != nil {
			return nil, err
		}
		row := &models.AffiliateInviteCode{
			Code:             code,
			Status:           constants.AffiliateInviteCodeStatusActive,
			CreatedByAdminID: input.CreatedByAdminID,
			Remark:           strings.TrimSpace(input.Remark),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.repo.CreateInviteCode(row); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		return s.repo.GetInviteCodeByID(row.ID)
	}
	return nil, ErrAffiliateCodeInvalid
}

// UpdateInviteCodeStatus 更新平台代理邀请码状态；已使用的邀请码不能重新启用。
func (s *AffiliateService) UpdateInviteCodeStatus(id uint, status string) (*models.AffiliateInviteCode, error) {
	if s.repo == nil || id == 0 {
		return nil, ErrNotFound
	}
	next := strings.TrimSpace(status)
	if next != constants.AffiliateInviteCodeStatusActive && next != constants.AffiliateInviteCodeStatusDisabled {
		return nil, ErrAffiliateInviteCodeInvalid
	}
	row, err := s.repo.GetInviteCodeByID(id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(row.Status) == constants.AffiliateInviteCodeStatusUsed {
		return nil, ErrAffiliateInviteCodeUnavailable
	}
	if strings.TrimSpace(row.Status) == next {
		return row, nil
	}
	if err := s.repo.UpdateInviteCodeStatus(id, next, time.Now()); err != nil {
		return nil, err
	}
	return s.repo.GetInviteCodeByID(id)
}

// CheckInviteCode 查询平台代理邀请码是否可用
func (s *AffiliateService) CheckInviteCode(code string) (bool, *models.AffiliateInviteCode, error) {
	if s.repo == nil {
		return false, nil, nil
	}
	row, err := s.repo.GetInviteCodeByCode(normalizeAffiliateInviteCode(code))
	if err != nil {
		return false, nil, err
	}
	if row == nil {
		return false, nil, nil
	}
	usable := strings.TrimSpace(row.Status) == constants.AffiliateInviteCodeStatusActive && row.UsedByUserID == nil && row.UsedAt == nil
	return usable, row, nil
}
