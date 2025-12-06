package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type InstanceRepository struct {
	instances sync.Map // map[string]*domain.Instance (Key: InstanceID)
}

func NewInstanceRepository() *InstanceRepository {
	return &InstanceRepository{}
}

func (r *InstanceRepository) Save(ctx context.Context, instance *model.Instance) error {
	r.instances.Store(instance.InstanceID, instance)
	return nil
}

func (r *InstanceRepository) FindByID(ctx context.Context, instanceID string) (*model.Instance, error) {
	val, ok := r.instances.Load(instanceID)
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}
	return val.(*model.Instance), nil
}

func (r *InstanceRepository) FindByUser(ctx context.Context, userID string) ([]*model.Instance, error) {
	var result []*model.Instance
	r.instances.Range(func(key, value any) bool {
		inst := value.(*model.Instance)
		if inst.UserID == userID {
			result = append(result, inst)
		}
		return true
	})
	return result, nil
}

func (r *InstanceRepository) Delete(ctx context.Context, instanceID string) error {
	r.instances.Delete(instanceID)
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

func (r *InstanceRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	count := 0
	r.instances.Range(func(key, value any) bool {
		inst := value.(*model.Instance)
		if inst.UserID == userID {
			count++
		}
		return true
	})
	return count, nil
}

func (r *InstanceRepository) CountByUserAndType(ctx context.Context, userID, instanceType string) (int, error) {
	count := 0
	r.instances.Range(func(key, value any) bool {
		inst := value.(*model.Instance)
		if inst.UserID == userID && inst.Type == instanceType {
			count++
		}
		return true
	})
	return count, nil
}