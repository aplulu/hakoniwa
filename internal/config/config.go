package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/kelseyhightower/envconfig"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type config struct {
	// Listen is the listen address.
	Listen string `envconfig:"LISTEN" default:""`
	// Port is the port number.
	Port string `envconfig:"PORT" default:"8080"`

	// KubeConfig is the path to the kubeconfig.
	KubeConfig string `envconfig:"KUBECONFIG" default:""`

	// KubernetesNamespace is the Kubernetes namespace to use.
	KubernetesNamespace string `envconfig:"KUBERNETES_NAMESPACE" default:"default"`

	// SwaggerUIEnabled is a flag to enable Swagger UI.
	SwaggerUIEnabled bool `envconfig:"SWAGGER_UI_ENABLED" default:"true"`

	// InstanceInactivityTimeout is the time duration after which an instance is considered inactive.
	InstanceInactivityTimeout time.Duration `envconfig:"INSTANCE_INACTIVITY_TIMEOUT" default:"1m"`

	// MaxPodCount is the maximum number of pods allowed (Global limit).
	MaxPodCount int `envconfig:"MAX_POD_COUNT" default:"100"`

	// MaxInstancesPerUser is the maximum number of instances allowed per user.
	MaxInstancesPerUser int `envconfig:"MAX_INSTANCES_PER_USER" default:"2"`

	// MaxInstancesPerUserPerType is the maximum number of instances allowed per user per type.
	MaxInstancesPerUserPerType int `envconfig:"MAX_INSTANCES_PER_USER_PER_TYPE" default:"1"`

	// PodTemplatePath is the path to the pod template file.
	PodTemplatePath string `envconfig:"POD_TEMPLATE_PATH" default:""`

	// Title is the application title.
	Title string `envconfig:"TITLE" default:"Hakoniwa"`

	// Message is the welcome message displayed below the title.
	Message string `envconfig:"MESSAGE" default:"On-Demand Cloud Workspace Environment"`

	// LogoURL is the URL to the application logo.
	LogoURL string `envconfig:"LOGO_URL" default:"/_hakoniwa/img/hakoniwa_logo.webp"`

	// TermsOfServiceURL is the URL to the terms of service.
	TermsOfServiceURL string `envconfig:"TERMS_OF_SERVICE_URL" default:""`

	// PrivacyPolicyURL is the URL to the privacy policy.
	PrivacyPolicyURL string `envconfig:"PRIVACY_POLICY_URL" default:""`

	// AuthMethods is the list of enabled authentication types.
	AuthMethods []string `envconfig:"AUTH_METHODS" default:"anonymous"`

	// AuthAutoLogin is a flag to automatically log in if only one auth method is enabled.
	AuthAutoLogin bool `envconfig:"AUTH_AUTO_LOGIN" default:"false"`

	// JWTSecret is the secret key for signing JWTs.
	JWTSecret string `envconfig:"JWT_SECRET" default:"hakoniwa-secret-key"`

	// SessionExpiration is the duration for which the session is valid.
	SessionExpiration time.Duration `envconfig:"SESSION_EXPIRATION" default:"24h"`

	// EnablePersistence is a flag to enable persistent storage globally.
	EnablePersistence bool `envconfig:"ENABLE_PERSISTENCE" default:"true"`

	// OIDCIssuerURL is the OIDC issuer URL.
	OIDCIssuerURL string `envconfig:"OIDC_ISSUER_URL" default:""`
	// OIDCClientID is the OpenID Connect client ID.
	OIDCClientID string `envconfig:"OIDC_CLIENT_ID" default:""`
	// OIDCClientSecret is the OpenID Connect client secret.
	OIDCClientSecret string `envconfig:"OIDC_CLIENT_SECRET" default:""`
	// OIDCRedirectURL is the OpenID Connect redirect URL.
	OIDCRedirectURL string `envconfig:"OIDC_REDIRECT_URL" default:""`
	// OIDCName is the display name for OpenID Connect login button.
	OIDCName string `envconfig:"OIDC_NAME" default:"OpenID Connect"`
	// OIDCScopes is the list of OpenID Connect scopes.
	OIDCScopes []string `envconfig:"OIDC_SCOPES" default:"openid,profile"`
}

type InstanceType struct {
	ID          string
	DisplayName string
	Description string
	LogoURL     string
	TargetPort  string // string to support named ports, though usually int
	Persistable bool
	Content     []byte
}

var (
	conf          config
	instanceTypes map[string]InstanceType
)

//go:embed pod_template.yaml
var defaultPodTemplate []byte

// LoadConf loads the configuration from the environment variables.
func LoadConf() error {
	if err := envconfig.Process("", &conf); err != nil {
		return fmt.Errorf("config.LoadConf: failed to load config: %w", err)
	}

	instanceTypes = make(map[string]InstanceType)

	var content []byte
	if conf.PodTemplatePath != "" {
		var err error
		content, err = os.ReadFile(conf.PodTemplatePath)
		if err != nil {
			return fmt.Errorf("failed to read pod template path %s: %w", conf.PodTemplatePath, err)
		}
	} else {
		content = defaultPodTemplate
	}

	// Decode multiple documents or List
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(content), 4096)
	for {
		// Verify if it is a generic map first to extract metadata
		var raw map[string]interface{}
		err := decoder.Decode(&raw)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode pod template: %w", err)
		}

		// Check if it's a List
		if kind, ok := raw["kind"].(string); ok && kind == "List" {
			if items, ok := raw["items"].([]interface{}); ok {
				for _, item := range items {
					// Re-marshal item to bytes to store in InstanceType
					// This is a bit inefficient but works
					// Or we can just parse manually.
					// Let's try to extract metadata from map.
					if itemMap, ok := item.(map[string]interface{}); ok {
						it, err := parseInstanceTypeMap(itemMap)
						if err != nil {
							return err
						}
						instanceTypes[it.ID] = it
					}
				}
			}
			continue
		}

		// Single Item (Pod)
		it, err := parseInstanceTypeMap(raw)
		if err != nil {
			return err
		}
		instanceTypes[it.ID] = it
	}

	// Fallback if empty?
	if len(instanceTypes) == 0 {
		// Should we error? Or just leave empty?
		// Original default was webtop.
	}

	return nil
}

func parseInstanceTypeMap(raw map[string]interface{}) (InstanceType, error) {
	// Extract metadata
	metadata, ok := raw["metadata"].(map[string]interface{})
	if !ok {
		return InstanceType{}, fmt.Errorf("missing metadata in pod template")
	}

	name, _ := metadata["name"].(string)
	if name == "" {
		return InstanceType{}, fmt.Errorf("missing metadata.name in pod template")
	}

	annotations, _ := metadata["annotations"].(map[string]interface{})

	displayName := name
	if val, ok := annotations["hakoniwa.aplulu.me/display-name"].(string); ok {
		displayName = val
	}

	description := ""
	if val, ok := annotations["hakoniwa.aplulu.me/description"].(string); ok {
		description = val
	}

	logoURL := ""
	if val, ok := annotations["hakoniwa.aplulu.me/image-url"].(string); ok {
		logoURL = val
	} else if val, ok := annotations["hakoniwa.aplulu.me/logo-url"].(string); ok {
		logoURL = val
	}

	targetPort := "3000"
	if val, ok := annotations["hakoniwa.aplulu.me/port"].(string); ok {
		targetPort = val
	}

	persistable := false
	if val, ok := annotations["hakoniwa.aplulu.me/persistable"].(string); ok {
		if val == "true" {
			persistable = true
		}
	}

	// Marshal back to bytes for Content
	// Note: This drops comments and re-formats, but that's acceptable for internal use.
	// We need a serializer. k8s yaml serializer?
	// Or just json marshal? K8s can handle JSON.
	// Let's use JSON marshaling as it's safer and built-in?
	// But we used yaml decoder.
	// Let's use "sigs.k8s.io/yaml" Marshal
	// We need to import it.
	// Wait, I removed sigs.k8s.io/yaml import in step 2 of previous change?
	// No, I kept it.
	// But here I used k8s.io/apimachinery/pkg/util/yaml.

	// Let's use a trick: we need the original bytes for this item.
	// Splitting by stream is hard to get original bytes.
	// Marshalling back is fine.

	content, err := yaml.Marshal(raw)
	if err != nil {
		return InstanceType{}, fmt.Errorf("failed to marshal item content: %w", err)
	}

	return InstanceType{
		ID:          name,
		DisplayName: displayName,
		Description: description,
		LogoURL:     logoURL,
		TargetPort:  targetPort,
		Persistable: persistable,
		Content:     content,
	}, nil
}

// GetInstanceType returns the instance type by ID.
func GetInstanceType(id string) (InstanceType, bool) {
	it, ok := instanceTypes[id]
	return it, ok
}

// GetInstanceTypes returns all available instance types.
func GetInstanceTypes() []InstanceType {
	types := make([]InstanceType, 0, len(instanceTypes))
	for _, it := range instanceTypes {
		types = append(types, it)
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i].DisplayName < types[j].DisplayName
	})
	return types
}

// MaxInstancesPerUser returns the max instances per user.
func MaxInstancesPerUser() int {
	return conf.MaxInstancesPerUser
}

// MaxInstancesPerUserPerType returns the max instances per user per type.
func MaxInstancesPerUserPerType() int {
	return conf.MaxInstancesPerUserPerType
}

// Listen returns the listen address.
func Listen() string {
	return conf.Listen
}

// Port returns the port number.
func Port() string {
	return conf.Port
}

// KubeConfig returns the path to the kubeconfig.
func KubeConfig() string {
	return conf.KubeConfig
}

// KubernetesNamespace returns the Kubernetes namespace to use.
func KubernetesNamespace() string {
	return conf.KubernetesNamespace
}

// SwaggerUIEnabled returns true if Swagger UI is enabled.
func SwaggerUIEnabled() bool {
	return conf.SwaggerUIEnabled
}

// InstanceInactivityTimeout returns the time duration after which an instance is considered inactive.
func InstanceInactivityTimeout() time.Duration {
	return conf.InstanceInactivityTimeout
}

// MaxPodCount returns the maximum number of pods allowed.
func MaxPodCount() int {
	return conf.MaxPodCount
}

// Title returns the application title.
func Title() string {
	return conf.Title
}

// Message returns the welcome message.
func Message() string {
	return conf.Message
}

// LogoURL returns the URL to the application logo.
func LogoURL() string {
	return conf.LogoURL
}

// TermsOfServiceURL returns the URL to the terms of service.
func TermsOfServiceURL() string {
	return conf.TermsOfServiceURL
}

// PrivacyPolicyURL returns the URL to the privacy policy.
func PrivacyPolicyURL() string {
	return conf.PrivacyPolicyURL
}

// AuthMethodsList returns the list of enabled authentication types.
func AuthMethodsList() []string {
	return conf.AuthMethods
}

// AuthAutoLogin returns true if automatic login should occur.
func AuthAutoLogin() bool {
	return conf.AuthAutoLogin
}

// JWTSecret returns the secret key for signing JWTs.
func JWTSecret() string {
	return conf.JWTSecret
}

// SessionExpiration returns the duration for which the session is valid.
func SessionExpiration() time.Duration {
	return conf.SessionExpiration
}

// EnablePersistence returns true if persistent storage is enabled globally.
func EnablePersistence() bool {
	return conf.EnablePersistence
}

// OIDCEnabled returns true if OIDC is enabled.
func OIDCEnabled() bool {
	for _, t := range conf.AuthMethods {
		if t == "oidc" {
			return true
		}
	}
	return false
}

// AnonymousEnabled returns true if anonymous login is enabled.
func AnonymousEnabled() bool {
	for _, t := range conf.AuthMethods {
		if t == "anonymous" {
			return true
		}
	}
	return false
}

// OIDCIssuerURL returns the OIDC issuer URL.
func OIDCIssuerURL() string {
	return conf.OIDCIssuerURL
}

// OIDCClientID returns the OIDC client ID.
func OIDCClientID() string {
	return conf.OIDCClientID
}

// OIDCClientSecret returns the OIDC client secret.
func OIDCClientSecret() string {
	return conf.OIDCClientSecret
}

// OIDCRedirectURL returns the OIDC redirect URL.
func OIDCRedirectURL() string {
	return conf.OIDCRedirectURL
}

// OIDCScopes returns the list of OIDC scopes.
func OIDCScopes() []string {
	return conf.OIDCScopes
}

// OIDCName returns the display name for OIDC login button.
func OIDCName() string {
	return conf.OIDCName
}
