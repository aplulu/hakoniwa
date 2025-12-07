package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/domain/model"
)

const (
	UserIDAnnotationKey = "hakoniwa.aplulu.com/user-id"
	ManagedByLabelKey   = "app.kubernetes.io/managed-by"
)

type Client struct {
	clientset *kubernetes.Clientset
	namespace string
	logger    *slog.Logger
}

func NewClient(logger *slog.Logger) (*Client, error) {
	var clusterConfig *rest.Config
	var err error
	if config.KubeConfig() != "" {
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", config.KubeConfig())
		if err != nil {
			return nil, fmt.Errorf("kubernetes.NewClient: failed to build config from kubeconfig: %w", err)
		}
	} else {
		clusterConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("kubernetes.NewClient: failed to build in-cluster config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("kubernetes.NewClient: failed to create clientset: %w", err)
	}

	// Get namespace from env or default
	namespace := config.KubernetesNamespace()
	if namespace == "" {
		namespace = "default"
	}

	return &Client{
		clientset: clientset,
		namespace: namespace,
		logger:    logger,
	}, nil
}

func (c *Client) CreateInstancePod(ctx context.Context, instance *model.Instance, templateContent []byte) error {
	if c.clientset == nil {
		return fmt.Errorf("kubernetes.CreateInstancePod: k8s client not configured (no-op mode)")
	}

	// Generate Pod Name: hakoniwa-{instance_id}
	// Assuming InstanceID is a UUID or safe string.
	podName := fmt.Sprintf("hakoniwa-%s", instance.InstanceID)
	instance.PodName = podName

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(templateContent), 4096)
	var u unstructured.Unstructured
	if err := decoder.Decode(&u); err != nil {
		return fmt.Errorf("kubernetes.CreateInstancePod: failed to decode pod template: %w", err)
	}

	u.SetName(podName)
	labels := u.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[ManagedByLabelKey] = "hakoniwa"
	// We keep user-id label for filtering/debugging, though instance-id is primary now
	sanitizedUser := sanitizeUserID(instance.UserID)
	labels["hakoniwa.aplulu.me/user-id"] = sanitizedUser
	u.SetLabels(labels)

	annotations := u.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[UserIDAnnotationKey] = instance.UserID
	annotations["hakoniwa.aplulu.me/instance-id"] = instance.InstanceID
	annotations["hakoniwa.aplulu.me/instance-type"] = instance.Type
	annotations["hakoniwa.aplulu.me/display-name"] = instance.DisplayName
	u.SetAnnotations(annotations)

	var pod corev1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &pod); err != nil {
		return fmt.Errorf("kubernetes.CreateInstancePod: failed to convert unstructured to pod: %w", err)
	}

	// Inject Environment Variables
	baseURL := "/"
	envVars := []corev1.EnvVar{
		{Name: "HAKONIWA_INSTANCE_ID", Value: instance.InstanceID},
		{Name: "HAKONIWA_BASE_URL", Value: baseURL},
	}

	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, envVars...)
	}

	_, err := c.clientset.CoreV1().Pods(c.namespace).Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("kubernetes.CreateInstancePod: failed to create pod: %w", err)
	}

	c.logger.Info("Created instance pod", "pod", podName, "user", instance.UserID, "type", instance.Type)
	return nil
}

func (c *Client) GetPodIP(ctx context.Context, podName string) (string, error) {
	if c.clientset == nil {
		return "", fmt.Errorf("kubernetes.CreateInstancePod: k8s client not configured (no-op mode)")
	}
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("kubernetes.CreateInstancePod: failed to get pod: %w", err)
	}

	if isPodReady(pod) {
		return pod.Status.PodIP, nil
	}

	return "", nil // Not running or not ready yet
}

func (c *Client) GetPodStatus(ctx context.Context, podName string) (model.InstanceStatus, string, error) {
	if c.clientset == nil {
		return "", "", fmt.Errorf("kubernetes.GetPodStatus: k8s client not configured")
	}
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return "", "", nil // Not Found
		}
		return "", "", fmt.Errorf("kubernetes.GetPodStatus: failed to get pod: %w", err)
	}

	// Check for termination/completion
	if pod.DeletionTimestamp != nil || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return model.InstanceStatusTerminating, "", nil
	}

	// Check Running & Ready
	if isPodReady(pod) {
		return model.InstanceStatusRunning, pod.Status.PodIP, nil
	}

	// Default to Pending (Phase is Pending, or Running but not Ready, or Unknown)
	return model.InstanceStatusPending, "", nil
}

func (c *Client) DeletePod(ctx context.Context, podName string) error {
	if c.clientset == nil {
		return fmt.Errorf("kubernetes.DeletePod: k8s client not configured (no-op mode)")
	}
	err := c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		// If not found, consider it deleted
		// Check error type if strict needed, but for now just return error
		return fmt.Errorf("kubernetes.DeletePod: failed to delete pod: %w", err)
	}
	c.logger.Info("Deleted instance pod", "pod", podName)
	return nil
}

func (c *Client) ListInstancePods(ctx context.Context) ([]*model.Instance, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("k8s client not configured (no-op mode)")
	}
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: ManagedByLabelKey + "=hakoniwa",
	})
	if err != nil {
		return nil, fmt.Errorf("kubernetes.ListInstancePods: failed to list pods: %w", err)
	}

	var instances []*model.Instance
	for _, pod := range pods.Items {
		// Skip terminating or finished pods
		if pod.DeletionTimestamp != nil || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}

		userID := pod.Annotations[UserIDAnnotationKey]
		instanceID := pod.Annotations["hakoniwa.aplulu.me/instance-id"]
		instanceType := pod.Annotations["hakoniwa.aplulu.me/instance-type"]
		displayName := pod.Annotations["hakoniwa.aplulu.me/display-name"]

		// Fallback for legacy pods
		if instanceID == "" {
			continue
		}

		status := model.InstanceStatusPending
		if isPodReady(&pod) {
			status = model.InstanceStatusRunning
		}
		// If not running (e.g. Pending, Unknown), it defaults to Pending.
		// Terminating (Succeeded/Failed) are filtered out above.

		// For recovery, we assume they are active now to prevent immediate cleanup
		lastActiveAt := time.Now()

		instances = append(instances, &model.Instance{
			InstanceID:   instanceID,
			UserID:       userID,
			Type:         instanceType,
			DisplayName:  displayName,
			PodName:      pod.Name,
			PodIP:        pod.Status.PodIP,
			Status:       status,
			LastActiveAt: lastActiveAt,
		})
	}
	return instances, nil
}

// sanitizeUserID makes the user ID safe for use in Kubernetes resource names
// (RFC 1123 subdomain label: lowercase alphanumeric, '-', start/end with alphanumeric)
func sanitizeUserID(userID string) string {
	// Replace invalid chars with '-'
	reg := regexp.MustCompile("[^a-z0-9]")
	lower := strings.ToLower(userID)
	safe := reg.ReplaceAllString(lower, "-")

	// Trim '-' from start and end
	safe = strings.Trim(safe, "-")

	// Max length 63 for label, but pod name also has prefix.
	// "hakoniwa-" is 9 chars. 63 - 9 = 54.
	if len(safe) > 54 {
		safe = safe[:54]
	}

	return safe
}

func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
