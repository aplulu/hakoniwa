package background

import (
	"context"
	"log/slog"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/repository"
)

type InactivityCleaner struct {
	instanceRepo repository.InstanceRepository
	k8sClient    repository.KubernetesClient
	logger       *slog.Logger
	timeout      time.Duration
}

func NewInactivityCleaner(
	instanceRepo repository.InstanceRepository,
	k8sClient repository.KubernetesClient,
	logger *slog.Logger,
	timeout time.Duration,
) *InactivityCleaner {
	return &InactivityCleaner{
		instanceRepo: instanceRepo,
		k8sClient:    k8sClient,
		logger:       logger,
		timeout:      timeout,
	}
}

func (c *InactivityCleaner) Start(ctx context.Context) {
	c.logger.Info("Starting inactivity cleaner", "timeout", c.timeout)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping inactivity cleaner")
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *InactivityCleaner) cleanup(ctx context.Context) {
	instances, err := c.instanceRepo.ListInactive(ctx, c.timeout)
	if err != nil {
		c.logger.Error("Failed to list inactive instances", "error", err)
		return
	}

	if len(instances) > 0 {
		c.logger.Info("Found inactive instances", "count", len(instances))
	}

	for _, instance := range instances {
		c.logger.Info("Cleaning up inactive instance", "user_id", instance.UserID, "pod_name", instance.PodName)

		// Delete Pod from K8s
		if err := c.k8sClient.DeletePod(ctx, instance.PodName); err != nil {
			c.logger.Error("Failed to delete pod", "pod_name", instance.PodName, "error", err)
			// If K8s deletion fails, skip repository deletion to retry later
			continue
		}

		// Delete from Repository
		if err := c.instanceRepo.Delete(ctx, instance.UserID); err != nil {
			c.logger.Error("Failed to delete instance from repository", "user_id", instance.UserID, "error", err)
		}
	}
}