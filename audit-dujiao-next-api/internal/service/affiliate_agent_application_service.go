package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

// AffiliateAgentApplicationCreateInput 创建平台代理申请输入。
type AffiliateAgentApplicationCreateInput struct {
	UserID  uint
	Phone   string
	Message string
}

// AffiliateAgentApplicationListInput 查询平台代理申请输入。
type AffiliateAgentApplicationListInput struct {
	Page     int
	PageSize int
	Status   string
	Keyword  string
}

// AffiliateAgentApplicationStatusInput 更新平台代理申请状态输入。
type AffiliateAgentApplicationStatusInput struct {
	Status    string
	AdminNote string
}

// CreateAgentApplication 创建平台代理申请。
func (s *AffiliateService) CreateAgentApplication(input AffiliateAgentApplicationCreateInput) (*models.AffiliateAgentApplication, error) {
	if s == nil || s.repo == nil || s.userRepo == nil || input.UserID == 0 {
		return nil, ErrAffiliateAgentApplicationInvalid
	}
	phone := cleanAgentApplicationText(input.Phone, 64)
	if phone == "" {
		return nil, ErrAffiliateAgentApplicationInvalid
	}
	user, err := s.userRepo.GetByID(input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	profile, err := s.repo.GetProfileByUserID(input.UserID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		return nil, ErrAffiliateAgentApplicationInvalid
	}
	latest, err := s.repo.GetLatestAgentApplicationByUserID(input.UserID)
	if err != nil {
		return nil, err
	}
	if latest != nil && strings.TrimSpace(latest.Status) == constants.AffiliateAgentApplicationStatusPending {
		return latest, ErrAffiliateAgentApplicationExists
	}

	now := time.Now()
	row := &models.AffiliateAgentApplication{
		UserID:    user.ID,
		Email:     user.Email,
		Phone:     phone,
		Message:   cleanAgentApplicationText(input.Message, 1000),
		Status:    constants.AffiliateAgentApplicationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateAgentApplication(row); err != nil {
		return nil, err
	}
	return s.repo.GetAgentApplicationByID(row.ID)
}

// GetLatestAgentApplicationByUserID 获取用户最新平台代理申请。
func (s *AffiliateService) GetLatestAgentApplicationByUserID(userID uint) (*models.AffiliateAgentApplication, error) {
	if s == nil || s.repo == nil || userID == 0 {
		return nil, nil
	}
	return s.repo.GetLatestAgentApplicationByUserID(userID)
}

// ListAgentApplications 查询平台代理申请。
func (s *AffiliateService) ListAgentApplications(input AffiliateAgentApplicationListInput) ([]models.AffiliateAgentApplication, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, ErrAffiliateAgentApplicationInvalid
	}
	return s.repo.ListAgentApplications(repository.AffiliateAgentApplicationListFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Status:   strings.TrimSpace(input.Status),
		Keyword:  strings.TrimSpace(input.Keyword),
	})
}

// UpdateAgentApplicationStatus 更新平台代理申请状态。
func (s *AffiliateService) UpdateAgentApplicationStatus(id uint, input AffiliateAgentApplicationStatusInput) (*models.AffiliateAgentApplication, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrAffiliateAgentApplicationInvalid
	}
	status := strings.TrimSpace(input.Status)
	if status != constants.AffiliateAgentApplicationStatusPending &&
		status != constants.AffiliateAgentApplicationStatusContacted &&
		status != constants.AffiliateAgentApplicationStatusRejected {
		return nil, ErrAffiliateAgentApplicationInvalid
	}
	row, err := s.repo.GetAgentApplicationByID(id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	if err := s.repo.UpdateAgentApplicationStatus(id, status, cleanAgentApplicationText(input.AdminNote, 255), time.Now()); err != nil {
		return nil, err
	}
	return s.repo.GetAgentApplicationByID(id)
}

func cleanAgentApplicationText(raw string, maxLen int) string {
	value := strings.TrimSpace(raw)
	if maxLen > 0 {
		runes := []rune(value)
		if len(runes) > maxLen {
			value = string(runes[:maxLen])
		}
	}
	return value
}
