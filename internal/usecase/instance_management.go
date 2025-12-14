package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/domain/model"
	"github.com/aplulu/hakoniwa/internal/domain/repository"
)

type InstanceManagement interface {
	ListInstances(ctx context.Context, userID string) ([]*model.Instance, error)
	GetInstance(ctx context.Context, instanceID string) (*model.Instance, error)
	CreateInstance(ctx context.Context, userID, instanceType string, persistent bool) (*model.Instance, error)
	DeleteInstance(ctx context.Context, userID, instanceID string) error
	UpdateLastActive(ctx context.Context, instanceID string) error
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

func (i *InstanceInteractor) UpdateLastActive(ctx context.Context, instanceID string) error {
	instance, err := i.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		return nil // Ignore if not found
	}
	instance.LastActiveAt = time.Now()
	return i.instanceRepo.Save(ctx, instance)
}

func (i *InstanceInteractor) ListInstances(ctx context.Context, userID string) ([]*model.Instance, error) {
	instances, err := i.instanceRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (i *InstanceInteractor) GetInstance(ctx context.Context, instanceID string) (*model.Instance, error) {
	instance, err := i.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Sync with K8s
	ip, err := i.k8sClient.GetPodIP(ctx, instance.PodName)
	if err != nil {
		return nil, err
	}

	newStatus := instance.Status
	if ip != "" {
		newStatus = model.InstanceStatusRunning
	} else {
		// If it was running but now no IP, maybe it crashed or is terminating?
		// Or just not ready?
		if instance.Status == model.InstanceStatusRunning {
			// Keep running if transient?
			// For now simple logic: No IP = Pending/Terminating
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

func (i *InstanceInteractor) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := i.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		return err
	}

	if instance.UserID != userID {
		return fmt.Errorf("instance not found") // Obfuscate
	}

	if err := i.k8sClient.DeletePod(ctx, instance.PodName); err != nil {
		// Log warning but continue to delete from repo?
		// Ideally we want consistency. If delete pod fails, maybe keep it?
		// But if pod is gone, we should delete from repo.
		// k8sClient.DeletePod should return nil if not found.
		return err
	}

	return i.instanceRepo.Delete(ctx, instanceID)
}

func (i *InstanceInteractor) CreateInstance(ctx context.Context, userID, instanceTypeID string, persistent bool) (*model.Instance, error) {
	// Enforce Global Persistence Setting
	if persistent && !config.EnablePersistence() {
		return nil, errors.New("persistent storage is disabled")
	}

	// Check Global Limit
	globalCount, err := i.instanceRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if globalCount >= config.MaxPodCount() {
		return nil, fmt.Errorf("max pod count reached")
	}

	// Check User Limit
	userCount, err := i.instanceRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userCount >= config.MaxInstancesPerUser() {
		return nil, fmt.Errorf("max instances per user reached")
	}

	// Check Type Limit
	typeCount, err := i.instanceRepo.CountByUserAndType(ctx, userID, instanceTypeID)
	if err != nil {
		return nil, err
	}
	if typeCount >= config.MaxInstancesPerUserPerType() {
		return nil, fmt.Errorf("max instances for this type reached")
	}

	// Get Template
	it, ok := config.GetInstanceType(instanceTypeID)
	if !ok {
		return nil, fmt.Errorf("invalid instance type: %s", instanceTypeID)
	}

	// Generate ID
	instanceID := uuid.New().String()

	instance := &model.Instance{
		InstanceID:   instanceID,
		UserID:       userID,
		Type:         instanceTypeID,
		DisplayName:  it.DisplayName,
		Status:       model.InstanceStatusPending,
		Persistent:   persistent,
		LastActiveAt: time.Now(),
		// PodName set by k8s client
	}

	if err := i.k8sClient.CreateInstancePod(ctx, instance, it.Content); err != nil {
		return nil, err
	}

	if err := i.instanceRepo.Save(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}