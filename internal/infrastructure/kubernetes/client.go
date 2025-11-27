package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/aplulu/hakoniwa/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
		return nil, fmt.Errorf("failed to create clientset: %w", err)
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

func (c *Client) CreateInstancePod(ctx context.Context, userID string) (string, error) {
	if c.clientset == nil {
		return "", fmt.Errorf("k8s client not configured (no-op mode)")
	}
	sanitizedID := sanitizeUserID(userID)
	podName := fmt.Sprintf("hakoniwa-%s", sanitizedID)

	templateBytes, err := config.GetPodTemplate(c.logger)
	if err != nil {
		return "", fmt.Errorf("failed to get pod template: %w", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(templateBytes), 4096)
	var u unstructured.Unstructured
	if err := decoder.Decode(&u); err != nil {
		return "", fmt.Errorf("failed to decode pod template: %w", err)
	}

	u.SetName(podName)
	labels := u.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["user"] = sanitizedID
	u.SetLabels(labels)

	var pod corev1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &pod); err != nil {
		return "", fmt.Errorf("failed to convert unstructured to pod: %w", err)
	}

	_, err = c.clientset.CoreV1().Pods(c.namespace).Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create pod: %w", err)
	}

	c.logger.Info("Created instance pod", "pod", podName, "user", userID)
	return podName, nil
}

func (c *Client) GetPodIP(ctx context.Context, podName string) (string, error) {
	if c.clientset == nil {
		return "", fmt.Errorf("k8s client not configured (no-op mode)")
	}
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %w", err)
	}

	if pod.Status.Phase == corev1.PodRunning {
		return pod.Status.PodIP, nil
	}

	return "", nil // Not running or no IP yet
}

func (c *Client) DeletePod(ctx context.Context, podName string) error {
	if c.clientset == nil {
		return fmt.Errorf("k8s client not configured (no-op mode)")
	}
	err := c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		// If not found, consider it deleted
		// Check error type if strict needed, but for now just return error
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	c.logger.Info("Deleted instance pod", "pod", podName)
	return nil
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
