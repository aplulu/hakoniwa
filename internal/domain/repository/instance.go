package repository

import (
	"context"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type InstanceRepository interface {
	Save(ctx context.Context, instance *model.Instance) error
	FindByID(ctx context.Context, instanceID string) (*model.Instance, error)
	FindByUser(ctx context.Context, userID string) ([]*model.Instance, error)
	Delete(ctx context.Context, instanceID string) error
	ListInactive(ctx context.Context, threshold time.Duration) ([]*model.Instance, error)
	Count(ctx context.Context) (int, error)
	CountByUser(ctx context.Context, userID string) (int, error)
	CountByUserAndType(ctx context.Context, userID, instanceType string) (int, error)
}