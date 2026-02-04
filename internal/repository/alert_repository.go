package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"gorm.io/gorm"
)

// AlertRepository handles log alert persistence
type AlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create inserts a new alert
func (r *AlertRepository) Create(ctx context.Context, alert *models.LogAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

// Update updates an alert
func (r *AlertRepository) Update(ctx context.Context, alert *models.LogAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

// FindByID retrieves an alert by ID
func (r *AlertRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.LogAlert, error) {
	var alert models.LogAlert
	err := r.db.WithContext(ctx).First(&alert, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// FindByTenantID retrieves all alerts for a tenant
func (r *AlertRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]models.LogAlert, error) {
	var alerts []models.LogAlert
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&alerts).Error
	return alerts, err
}

// FindEnabled retrieves all enabled alerts
func (r *AlertRepository) FindEnabled(ctx context.Context) ([]models.LogAlert, error) {
	var alerts []models.LogAlert
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&alerts).Error
	return alerts, err
}

// Delete removes an alert
func (r *AlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.LogAlert{}, "id = ?", id).Error
}

// UpdateLastTriggered updates the last triggered timestamp
func (r *AlertRepository) UpdateLastTriggered(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.LogAlert{}).
		Where("id = ?", id).
		Update("last_triggered", gorm.Expr("NOW()")).Error
}
