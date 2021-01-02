package middleware

import (
	"context"
	"net/http"
	"strings"

	core_auth_sdk "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-auth-sdk"
	core_logging "github.com/BlackspaceInc/BlackspacePlatform/src/libraries/core/core-logging/json"
	"github.com/keratin/authn-go/authn"
)

var UserCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type AuthnMW struct {
	client *core_auth_sdk.Client
	logger core_logging.ILog
}

func NewAuthnMw(c *core_auth_sdk.Client, log core_logging.ILog) *AuthnMW {
	return &AuthnMW{client: c, logger: log}
}

func (mw *AuthnMW) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()
		authorization := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authorization, "Bearer ")
		mw.logger.InfoM(token)
		decodedToken, err := mw.client.SubjectFrom(token)
		mw.logger.Error(err, "error")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(ctx, UserCtxKey, decodedToken)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// AuthnMiddleware decodes the jwt token and packs the jwt into the context.
func AuthnMiddleware(auth *authn.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			w.Header().Set("Content-Type", "application/json")

			ctx := r.Context()
			authorization := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authorization, "Bearer ")
			decodedToken, err := auth.SubjectFrom(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx = context.WithValue(ctx, UserCtxKey, decodedToken)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// IsAuthenticated returns wether or not the user is authenticated.
// REQUIRES Middleware to have run.
func IsAuthenticated(ctx context.Context) bool {
	return ctx.Value(UserCtxKey) != nil
}
