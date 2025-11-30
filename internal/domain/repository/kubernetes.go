package repository

import (
	"context"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type KubernetesClient interface {
	CreateInstancePod(ctx context.Context, userID string) (string, error)
	GetPodIP(ctx context.Context, podName string) (string, error)
	DeletePod(ctx context.Context, podName string) error
	ListInstancePods(ctx context.Context) ([]*model.Instance, error)
}
