package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/repository"
)

// AlertService handles alert business logic
type AlertService struct {
	repo *repository.AlertRepository
}

// NewAlertService creates a new alert service
func NewAlertService(repo *repository.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(ctx context.Context, alert *models.LogAlert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	return s.repo.Create(ctx, alert)
}

// UpdateAlert updates an alert
func (s *AlertService) UpdateAlert(ctx context.Context, alert *models.LogAlert) error {
	return s.repo.Update(ctx, alert)
}

// GetAlert retrieves an alert by ID
func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*models.LogAlert, error) {
	return s.repo.FindByID(ctx, id)
}

// GetAlertsByTenant retrieves all alerts for a tenant
func (s *AlertService) GetAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.LogAlert, error) {
	return s.repo.FindByTenantID(ctx, tenantID)
}

// DeleteAlert removes an alert
func (s *AlertService) DeleteAlert(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// EnableAlert enables an alert
func (s *AlertService) EnableAlert(ctx context.Context, id uuid.UUID) error {
	alert, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	alert.Enabled = true
	return s.repo.Update(ctx, alert)
}

// DisableAlert disables an alert
func (s *AlertService) DisableAlert(ctx context.Context, id uuid.UUID) error {
	alert, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	alert.Enabled = false
	return s.repo.Update(ctx, alert)
}

// GetEnabledAlerts retrieves all enabled alerts
func (s *AlertService) GetEnabledAlerts(ctx context.Context) ([]models.LogAlert, error) {
	return s.repo.FindEnabled(ctx)
}
