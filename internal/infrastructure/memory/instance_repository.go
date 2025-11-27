package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type InstanceRepository struct {
	instances sync.Map // map[string]*domain.Instance (Key: UserID)
}

func NewInstanceRepository() *InstanceRepository {
	return &InstanceRepository{}
}

func (r *InstanceRepository) Save(ctx context.Context, instance *model.Instance) error {
	r.instances.Store(instance.UserID, instance)
	return nil
}

func (r *InstanceRepository) Get(ctx context.Context, userID string) (*model.Instance, error) {
	val, ok := r.instances.Load(userID)
	if !ok {
		return nil, fmt.Errorf("instance not found for user: %s", userID)
	}
	return val.(*model.Instance), nil
}

func (r *InstanceRepository) Delete(ctx context.Context, userID string) error {
	r.instances.Delete(userID)
	return nil
}

func (r *InstanceRepository) ListInactive(ctx context.Context, threshold time.Duration) ([]*model.Instance, error) {
	var inactiveInstances []*model.Instance
	now := time.Now()

	r.instances.Range(func(key, value any) bool {
		instance := value.(*model.Instance)
		// If LastActiveAt is older than the threshold, it's inactive.
		if now.Sub(instance.LastActiveAt) > threshold {
			inactiveInstances = append(inactiveInstances, instance)
		}
		return true
	})

	return inactiveInstances, nil
}

func (r *InstanceRepository) Count(ctx context.Context) (int, error) {
	count := 0
	r.instances.Range(func(key, value any) bool {
		count++
		return true
	})
	return count, nil
}
