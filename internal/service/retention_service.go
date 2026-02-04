package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/repository"
)

// RetentionService handles retention policy business logic
type RetentionService struct {
	repo *repository.RetentionRepository
}

// NewRetentionService creates a new retention service
func NewRetentionService(repo *repository.RetentionRepository) *RetentionService {
	return &RetentionService{repo: repo}
}

// CreatePolicy creates a new retention policy
func (s *RetentionService) CreatePolicy(ctx context.Context, policy *models.LogRetention) error {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	return s.repo.Create(ctx, policy)
}

// UpdatePolicy updates a retention policy
func (s *RetentionService) UpdatePolicy(ctx context.Context, policy *models.LogRetention) error {
	return s.repo.Update(ctx, policy)
}

// GetPolicy retrieves retention policy for a tenant
func (s *RetentionService) GetPolicy(ctx context.Context, tenantID uuid.UUID) (*models.LogRetention, error) {
	return s.repo.FindByTenantID(ctx, tenantID)
}

// GetAllPolicies retrieves all retention policies
func (s *RetentionService) GetAllPolicies(ctx context.Context) ([]models.LogRetention, error) {
	return s.repo.FindAll(ctx)
}

// DeletePolicy removes a retention policy
func (s *RetentionService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// UpsertPolicy creates or updates a retention policy
func (s *RetentionService) UpsertPolicy(ctx context.Context, policy *models.LogRetention) error {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	return s.repo.Upsert(ctx, policy)
}
