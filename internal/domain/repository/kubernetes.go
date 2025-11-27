package repository

import "context"

type KubernetesClient interface {
	CreateInstancePod(ctx context.Context, userID string) (string, error)
	GetPodIP(ctx context.Context, podName string) (string, error)
	DeletePod(ctx context.Context, podName string) error
}
