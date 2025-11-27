package usecase

import (
	"context"
	"time"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/domain/model"
	"github.com/aplulu/hakoniwa/internal/domain/repository"
)

type InstanceManagement interface {
	GetInstanceStatus(ctx context.Context, userID string) (*model.Instance, error)
	CreateInstance(ctx context.Context, userID string) error
}

type InstanceInteractor struct {
	instanceRepo repository.InstanceRepository
	k8sClient    repository.KubernetesClient
}

func NewInstanceInteractor(instanceRepo repository.InstanceRepository, k8sClient repository.KubernetesClient) InstanceManagement {
	return &InstanceInteractor{
		instanceRepo: instanceRepo,
		k8sClient:    k8sClient,
	}
}

func (i *InstanceInteractor) GetInstanceStatus(ctx context.Context, userID string) (*model.Instance, error) {
	instance, err := i.instanceRepo.Get(ctx, userID)
	if err != nil {
		// If not found, it's not an error, just return nil
		// But Get returns error if not found based on implementation?
		// Let's assume we return nil, nil if not found for the handler to handle
		// Current memory repo returns error "instance not found".
		// We should handle that.
		// Let's just bubble up the error for now, handler can check if it's "not found" or use a specific error type.
		// But here we want to return nil if not found.
		return nil, nil
	}

	ip, err := i.k8sClient.GetPodIP(ctx, instance.PodName)
	if err != nil {
		// If error accessing K8s, keep last known status? Or return error?
		return nil, err
	}

	newStatus := instance.Status
	if ip != "" {
		newStatus = model.InstanceStatusRunning
	} else {
		if instance.Status == model.InstanceStatusRunning {
			newStatus = model.InstanceStatusPending
		}
	}

	instance.Status = newStatus
	instance.PodIP = ip
	instance.LastActiveAt = time.Now()

	if err := i.instanceRepo.Save(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

func (i *InstanceInteractor) CreateInstance(ctx context.Context, userID string) error {
	existing, _ := i.instanceRepo.Get(ctx, userID)
	if existing != nil {
		// Already exists, maybe just ensure it's running?
		// For now, assume create is called only when needed.
		return nil
	}

	// Check max instances
	count, err := i.instanceRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count >= config.MaxPodCount() {
		return model.ErrMaxInstancesReached
	}

	podName, err := i.k8sClient.CreateInstancePod(ctx, userID)
	if err != nil {
		return err
	}

	// Create Instance Record
	instance := &model.Instance{
		UserID:       userID,
		PodName:      podName,
		Status:       model.InstanceStatusPending,
		LastActiveAt: time.Now(),
	}

	if err := i.instanceRepo.Save(ctx, instance); err != nil {
		return err
	}

	return nil
}
