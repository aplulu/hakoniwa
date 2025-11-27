package repository

import (
	"context"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type InstanceRepository interface {
	Save(ctx context.Context, instance *model.Instance) error
	Get(ctx context.Context, userID string) (*model.Instance, error)
	Delete(ctx context.Context, userID string) error
	ListInactive(ctx context.Context, threshold time.Duration) ([]*model.Instance, error)
	Count(ctx context.Context) (int, error)
}
