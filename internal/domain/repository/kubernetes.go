package repository

import (
	"context"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type KubernetesClient interface {

	CreateInstancePod(ctx context.Context, instance *model.Instance, template []byte) error

	GetPodIP(ctx context.Context, podName string) (string, error)

	GetPodStatus(ctx context.Context, podName string) (model.InstanceStatus, string, error)

	DeletePod(ctx context.Context, podName string) error

	ListInstancePods(ctx context.Context) ([]*model.Instance, error)

}
