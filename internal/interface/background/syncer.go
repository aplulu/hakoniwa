package background

import (
	"context"
	"log/slog"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/repository"
)

type InstanceSyncer struct {
	instanceRepo repository.InstanceRepository
	k8sClient    repository.KubernetesClient
	logger       *slog.Logger
}

func NewInstanceSyncer(
	instanceRepo repository.InstanceRepository,
	k8sClient repository.KubernetesClient,
	logger *slog.Logger,
) *InstanceSyncer {
	return &InstanceSyncer{
		instanceRepo: instanceRepo,
		k8sClient:    k8sClient,
		logger:       logger,
	}
}

func (s *InstanceSyncer) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting instance syncer", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping instance syncer")
			return
		case <-ticker.C:
			if err := s.sync(ctx); err != nil {
				s.logger.Error("Failed to sync instances", "error", err)
			}
		}
	}
}

func (s *InstanceSyncer) sync(ctx context.Context) error {
	// 1. Get all instances from K8s
	k8sInstances, err := s.k8sClient.ListInstancePods(ctx)
	if err != nil {
		return err
	}

	// Map for easy lookup
	k8sMap := make(map[string]struct{})
	for _, inst := range k8sInstances {
		k8sMap[inst.InstanceID] = struct{}{}
		
		// 2. Update or Add to Repo
		existing, err := s.instanceRepo.FindByID(ctx, inst.InstanceID)
		if err != nil {
			// Not found in repo -> Add (Recovery)
			// Log?
			// s.logger.Info("Recovering instance from K8s", "id", inst.InstanceID)
			if err := s.instanceRepo.Save(ctx, inst); err != nil {
				s.logger.Error("Failed to save recovered instance", "id", inst.InstanceID, "error", err)
			}
		} else {
			// Found -> Update Status and IP only
			existing.Status = inst.Status
			existing.PodIP = inst.PodIP
			// existing.LastActiveAt is PRESERVED
			if err := s.instanceRepo.Save(ctx, existing); err != nil {
				s.logger.Error("Failed to update instance", "id", inst.InstanceID, "error", err)
			}
		}
	}

	// 3. Remove instances from Repo that are not in K8s
	// Note: ListInstancePods excludes Terminated/Failed/Succeeded.
	// So if it's in Repo but not in k8sInstances, it's effectively dead/gone.
	
	// We need a way to iterate ALL instances in repo.
	// InstanceRepository.Count/List?
	// The current interface doesn't have "ListAll".
	// It has `ListInactive`.
	// I should add `ListAll` or `Iterate` to Repository interface?
	// Memory implementation has `Range`.
	// Let's add `ListAll(ctx)` to repository interface.
	
	// Wait, `ListInactive(0)` returns all?
	// `now.Sub(instance.LastActiveAt) > threshold`.
	// If threshold is -1h (negative), `now - lastActive > -1h` -> always true?
	// Yes.
	allInstances, err := s.instanceRepo.ListInactive(ctx, -10000*time.Hour)
	if err != nil {
		return err
	}

	for _, repoInst := range allInstances {
		if _, ok := k8sMap[repoInst.InstanceID]; !ok {
			// Not in K8s list -> Delete
			// s.logger.Info("Removing missing instance from repo", "id", repoInst.InstanceID)
			if err := s.instanceRepo.Delete(ctx, repoInst.InstanceID); err != nil {
				s.logger.Error("Failed to delete missing instance", "id", repoInst.InstanceID, "error", err)
			}
		}
	}

	return nil
}
