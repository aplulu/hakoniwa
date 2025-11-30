package config

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
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

	// MaxPodCount is the maximum number of pods allowed.
	MaxPodCount int `envconfig:"MAX_POD_COUNT" default:"3"`

	// PodTemplatePath is the path to the pod template file.
	PodTemplatePath string `envconfig:"POD_TEMPLATE_PATH" default:""`

	// Title is the application title.
	Title string `envconfig:"TITLE" default:"Hakoniwa"`

	// Message is the welcome message displayed below the title.
	Message string `envconfig:"MESSAGE" default:"On-Demand Cloud Desktop Environment"`

	// LogoURL is the URL to the application logo.
	LogoURL string `envconfig:"LOGO_URL" default:"/_hakoniwa/hakoniwa_logo.webp"`

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

var conf config

//go:embed pod_template.yaml
var defaultPodTemplate []byte

// LoadConf loads the configuration from the environment variables.
func LoadConf() error {
	if err := envconfig.Process("", &conf); err != nil {
		return fmt.Errorf("config.LoadConf: failed to load config: %w", err)
	}

	return nil
}

// GetPodTemplate returns the pod template as bytes.
// It checks the configured PodTemplatePath. If it exists, it reads from there.
// Otherwise, it returns the default embedded template.
func GetPodTemplate(logger *slog.Logger) ([]byte, error) {
	path := PodTemplatePath()
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return os.ReadFile(path)
		}
		if logger != nil {
			logger.Warn("Pod template path specified but file not found, using default", "path", path)
		}
	}
	return defaultPodTemplate, nil
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

// PodTemplatePath returns the path to the pod template file.
func PodTemplatePath() string {
	return conf.PodTemplatePath
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
