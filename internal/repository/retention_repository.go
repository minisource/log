package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"gorm.io/gorm"
)

// RetentionRepository handles retention policy persistence
type RetentionRepository struct {
	db *gorm.DB
}

// NewRetentionRepository creates a new retention repository
func NewRetentionRepository(db *gorm.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// Create inserts a new retention policy
func (r *RetentionRepository) Create(ctx context.Context, retention *models.LogRetention) error {
	return r.db.WithContext(ctx).Create(retention).Error
}

// Update updates a retention policy
func (r *RetentionRepository) Update(ctx context.Context, retention *models.LogRetention) error {
	return r.db.WithContext(ctx).Save(retention).Error
}

// FindByTenantID retrieves retention policy for a tenant
func (r *RetentionRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.LogRetention, error) {
	var retention models.LogRetention
	err := r.db.WithContext(ctx).First(&retention, "tenant_id = ?", tenantID).Error
	if err != nil {
		return nil, err
	}
	return &retention, nil
}

// FindAll retrieves all retention policies
func (r *RetentionRepository) FindAll(ctx context.Context) ([]models.LogRetention, error) {
	var policies []models.LogRetention
	err := r.db.WithContext(ctx).Find(&policies).Error
	return policies, err
}

// Delete removes a retention policy
func (r *RetentionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.LogRetention{}, "id = ?", id).Error
}

// Upsert creates or updates a retention policy
func (r *RetentionRepository) Upsert(ctx context.Context, retention *models.LogRetention) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ?", retention.TenantID).
		Assign(retention).
		FirstOrCreate(retention).Error
}
