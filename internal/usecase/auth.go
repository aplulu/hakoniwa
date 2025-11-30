package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/domain/model"
)

type Auth interface {
	VerifySession(ctx context.Context, token string) (*model.User, string, error)
	LoginAnonymous(ctx context.Context) (string, *model.User, error) // returns token, user, error
	LoginOIDC(ctx context.Context, redirectURI string) (string, error)
	CallbackOIDC(ctx context.Context, code string, state string, redirectURI string) (string, *model.User, error)
}

type AuthInteractor struct {
	oidcProvider *oidc.Provider
	oidcVerifier *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
}

type CustomClaims struct {
	UserID   string `json:"user_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

func NewAuthInteractor() (*AuthInteractor, error) {
	ai := &AuthInteractor{}

	if config.OIDCEnabled() {
		provider, err := oidc.NewProvider(context.Background(), config.OIDCIssuerURL())
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}
		ai.oidcProvider = provider

		ai.oidcVerifier = provider.Verifier(&oidc.Config{
			ClientID: config.OIDCClientID(),
		})

		ai.oauth2Config = &oauth2.Config{
			ClientID:     config.OIDCClientID(),
			ClientSecret: config.OIDCClientSecret(),
			Endpoint:     provider.Endpoint(),
			RedirectURL:  config.OIDCRedirectURL(),
			Scopes:       config.OIDCScopes(),
		}
	}

	return ai, nil
}

func (a *AuthInteractor) VerifySession(ctx context.Context, tokenString string) (*model.User, string, error) {
	if tokenString == "" {
		return nil, "", model.ErrUnauthorized
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret()), nil
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		user := &model.User{
			ID:   claims.UserID,
			Type: model.UserType(claims.UserType),
		}

		// Sliding Session Check
		newToken := ""
		expiration := claims.ExpiresAt.Time
		now := time.Now()
		sessionDuration := config.SessionExpiration()
		
		// If remaining time is less than half of session duration, renew token
		if expiration.Sub(now) < sessionDuration/2 {
			nt, err := a.createToken(user)
			if err == nil {
				newToken = nt
			} else {
				// Log warning?
			}
		}

		return user, newToken, nil
	}

	return nil, "", model.ErrUnauthorized
}

func (a *AuthInteractor) LoginAnonymous(ctx context.Context) (string, *model.User, error) {
	uuidObj := uuid.New()
	id := "anon-" + uuidObj.String()

	user := &model.User{
		ID:   id,
		Type: model.UserTypeAnonymous,
	}

	token, err := a.createToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (a *AuthInteractor) LoginOIDC(ctx context.Context, redirectURI string) (string, error) {
	if a.oauth2Config == nil {
		return "", errors.New("OIDC is not enabled")
	}

	state, err := generateRandomState()
	if err != nil {
		return "", err
	}

	// Copy config to support dynamic redirect URI if provided
	conf := *a.oauth2Config
	if redirectURI != "" {
		conf.RedirectURL = redirectURI
	}

	return conf.AuthCodeURL(state), nil
}

func (a *AuthInteractor) CallbackOIDC(ctx context.Context, code string, state string, redirectURI string) (string, *model.User, error) {
	if a.oauth2Config == nil {
		return "", nil, errors.New("OIDC is not enabled")
	}

	// Verify state? (Skipped for now, relies on frontend/state matching)

	conf := *a.oauth2Config
	if redirectURI != "" {
		conf.RedirectURL = redirectURI
	}

	oauth2Token, err := conf.Exchange(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", nil, errors.New("no id_token field in oauth2 token")
	}

	idToken, err := a.oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	var claims struct {
		Subject string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	user := &model.User{
		ID:   "oidc:" + claims.Subject,
		Type: model.UserTypeOIDC,
	}

	token, err := a.createToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (a *AuthInteractor) createToken(user *model.User) (string, error) {
	claims := CustomClaims{
		user.ID,
		string(user.Type),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.SessionExpiration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "hakoniwa",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(config.JWTSecret()))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return ss, nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
