package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/aplulu/hakoniwa/internal/api/hakoniwa"
	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/interface/http/middleware"
	"github.com/aplulu/hakoniwa/internal/usecase"
)

type APIHandler struct {
	authUsecase     usecase.Auth
	instanceUsecase usecase.InstanceManagement
}

func NewAPIHandler(auth usecase.Auth, instance usecase.InstanceManagement) *APIHandler {
	return &APIHandler{
		authUsecase:     auth,
		instanceUsecase: instance,
	}
}

// GetAuthMe implements getAuthMe operation.
// GET /auth/me
func (h *APIHandler) GetAuthMe(ctx context.Context) (hakoniwa.GetAuthMeRes, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return &hakoniwa.GetAuthMeUnauthorized{}, nil
	}

	res := &hakoniwa.AuthStatus{
		User: hakoniwa.User{
			ID:   user.ID,
			Type: hakoniwa.UserType(user.Type),
		},
	}

	return res, nil
}

// LoginAnonymous implements loginAnonymous operation.
// POST /auth/anonymous
func (h *APIHandler) LoginAnonymous(ctx context.Context) (*hakoniwa.AuthStatus, error) {
	token, user, err := h.authUsecase.LoginAnonymous(ctx)
	if err != nil {
		return nil, err
	}

	if setter, ok := ctx.Value(CookieSetterKey).(func(string)); ok {
		setter(token)
	}

	return &hakoniwa.AuthStatus{
		User: hakoniwa.User{
			ID:   user.ID,
			Type: hakoniwa.UserType(user.Type),
		},
	}, nil
}

// OidcAuthorize implements oidcAuthorize operation.
// GET /auth/oidc/authorize
func (h *APIHandler) OidcAuthorize(ctx context.Context) (*hakoniwa.OidcAuthorizeFound, error) {
	url, err := h.authUsecase.LoginOIDC(ctx, config.OIDCRedirectURL())
	if err != nil {
		return nil, err
	}
	return &hakoniwa.OidcAuthorizeFound{
		Location: url,
	}, nil
}

// OidcCallback implements oidcCallback operation.
// GET /auth/oidc/callback
func (h *APIHandler) OidcCallback(ctx context.Context, params hakoniwa.OidcCallbackParams) (*hakoniwa.OidcCallbackFound, error) {
	// Handle IdP errors
	if params.Error.IsSet() {
		return &hakoniwa.OidcCallbackFound{
			Location: "/?error=" + params.Error.Value,
		}, nil
	}

	// Use configured redirect URI
	redirectURI := config.OIDCRedirectURL()

	token, _, err := h.authUsecase.CallbackOIDC(ctx, params.Code, params.State, redirectURI)
	if err != nil {
		return &hakoniwa.OidcCallbackFound{
			Location: "/?error=login_failed",
		}, nil
	}

	// Set session cookie
	if setter, ok := ctx.Value(CookieSetterKey).(func(string)); ok {
		setter(token)
	}

	// Redirect to dashboard
	return &hakoniwa.OidcCallbackFound{
		Location: "/",
	}, nil
}

// ListInstances implements listInstances operation.
// GET /instances
func (h *APIHandler) ListInstances(ctx context.Context) ([]hakoniwa.Instance, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	instances, err := h.instanceUsecase.ListInstances(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	res := make([]hakoniwa.Instance, 0, len(instances))
	for _, inst := range instances {
		res = append(res, hakoniwa.Instance{
			ID:     inst.InstanceID,
			Name:   inst.DisplayName,
			Type:   inst.Type,
			Status: hakoniwa.InstanceStatus(inst.Status),
			PodIP:  hakoniwa.NewOptString(inst.PodIP),
		})
	}

	return res, nil
}

// CreateInstance implements createInstance operation.
// POST /instances
func (h *APIHandler) CreateInstance(ctx context.Context, req *hakoniwa.CreateInstanceRequest) (hakoniwa.CreateInstanceRes, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	inst, err := h.instanceUsecase.CreateInstance(ctx, user.ID, req.Type)
	if err != nil {
		// Check for specific errors
		if err.Error() == "max pod count reached" || err.Error() == "max instances per user reached" || err.Error() == "max instances for this type reached" {
			return &hakoniwa.CreateInstanceServiceUnavailable{}, nil
		}
		// Assuming invalid type returns 400?
		if err.Error() == fmt.Sprintf("invalid instance type: %s", req.Type) {
			return &hakoniwa.CreateInstanceBadRequest{}, nil
		}

		return nil, err
	}

	return &hakoniwa.Instance{
		ID:     inst.InstanceID,
		Name:   inst.DisplayName,
		Type:   inst.Type,
		Status: hakoniwa.InstanceStatus(inst.Status),
		PodIP:  hakoniwa.NewOptString(inst.PodIP),
	}, nil
}

// DeleteInstance implements deleteInstance operation.
// DELETE /instances/{instanceId}
func (h *APIHandler) DeleteInstance(ctx context.Context, params hakoniwa.DeleteInstanceParams) (hakoniwa.DeleteInstanceRes, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	err := h.instanceUsecase.DeleteInstance(ctx, user.ID, params.InstanceId)
	if err != nil {
		// If instance not found or not owned by user
		if err.Error() == "instance not found" {
			return &hakoniwa.DeleteInstanceNotFound{}, nil
		}
		return nil, err
	}

	return &hakoniwa.DeleteInstanceNoContent{}, nil
}

// ListInstanceTypes implements listInstanceTypes operation.
// GET /instance-types
func (h *APIHandler) ListInstanceTypes(ctx context.Context) ([]hakoniwa.InstanceType, error) {
	types := config.GetInstanceTypes()
	res := make([]hakoniwa.InstanceType, 0, len(types))
	for _, t := range types {
		res = append(res, hakoniwa.InstanceType{
			ID:          t.ID,
			Name:        t.DisplayName,
			Description: hakoniwa.NewOptString(t.Description),
			LogoURL:     hakoniwa.NewOptString(t.LogoURL),
		})
	}
	return res, nil
}

// GetConfiguration implements getConfiguration operation.
// GET /configuration
func (h *APIHandler) GetConfiguration(ctx context.Context) (*hakoniwa.Configuration, error) {
	return &hakoniwa.Configuration{
		Title:             config.Title(),
		Message:           config.Message(),
		LogoURL:           config.LogoURL(),
		TermsOfServiceURL: hakoniwa.NewOptString(config.TermsOfServiceURL()),
		PrivacyPolicyURL:  hakoniwa.NewOptString(config.PrivacyPolicyURL()),
		AuthMethods:       config.AuthMethodsList(),
		OidcName:          hakoniwa.NewOptString(config.OIDCName()),
		AuthAutoLogin:     config.AuthAutoLogin(),
	}, nil
}

// Logout implements logout operation.

// POST /auth/logout

func (h *APIHandler) Logout(ctx context.Context) error {

	if clearer, ok := ctx.Value(CookieClearerKey).(func()); ok {

		clearer()

	}

	return nil

}



// CookieSetterKey is used to inject a callback to set cookies

type cookieSetterKey struct{}



var CookieSetterKey = cookieSetterKey{}



// CookieClearerKey is used to inject a callback to clear cookies

type cookieClearerKey struct{}



var CookieClearerKey = cookieClearerKey{}



func WithCookieSetter(ctx context.Context, setter func(token string)) context.Context {

	return context.WithValue(ctx, CookieSetterKey, setter)

}



func WithCookieClearer(ctx context.Context, clearer func()) context.Context {

	return context.WithValue(ctx, CookieClearerKey, clearer)

}
