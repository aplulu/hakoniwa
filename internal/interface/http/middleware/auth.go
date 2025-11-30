package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/domain/model"
	"github.com/aplulu/hakoniwa/internal/usecase"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type AuthMiddleware struct {
	authUsecase usecase.Auth
}

func NewAuthMiddleware(authUsecase usecase.Auth) *AuthMiddleware {
	return &AuthMiddleware{
		authUsecase: authUsecase,
	}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("hakoniwa_session")
		if err != nil {
			// No session cookie, proceed as anonymous (user not in context)
			next.ServeHTTP(w, r)
			return
		}

		user, newToken, err := m.authUsecase.VerifySession(r.Context(), cookie.Value)
		if err != nil {
			// Invalid session, proceed without user
			// Optionally clear cookie here?
			next.ServeHTTP(w, r)
			return
		}

		// If a new token was issued (sliding session), set it in the cookie
		if newToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "hakoniwa_session",
				Value:    newToken,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(config.SessionExpiration()),
			})
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*model.User)
	return user, ok
}
