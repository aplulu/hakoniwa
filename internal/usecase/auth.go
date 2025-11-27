package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type Auth interface {
	VerifySession(ctx context.Context, token string) (*model.User, error)
	LoginAnonymous(ctx context.Context) (string, *model.User, error) // returns token, user, error
}

type AuthInteractor struct {
	// In future, user repository or OIDC provider
}

func NewAuthInteractor() Auth {
	return &AuthInteractor{}
}

func (a *AuthInteractor) VerifySession(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, model.ErrUnauthorized
	}

	// Detect type
	userType := model.UserTypeAnonymous
	if strings.HasPrefix(token, "oidc:") {
		userType = model.UserTypeOIDC
	}

	return &model.User{
		ID:   token,
		Type: userType,
	}, nil
}

func (a *AuthInteractor) LoginAnonymous(ctx context.Context) (string, *model.User, error) {
	uuidObj := uuid.New()
	id := "anon-" + uuidObj.String()

	user := &model.User{
		ID:   id,
		Type: model.UserTypeAnonymous,
	}

	token := id

	return token, user, nil
}
