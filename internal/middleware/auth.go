package middleware

import (
	"errors"
	"net/http"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
)

func Auth(cfg config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			l := logging.Logger
			cookieToken, err := r.Cookie("auth_token")
			var tokenString string
			var userID string
			newTokenRequired := false

			if err == nil {
				tokenString = cookieToken.Value
				userID, err = auth.ValidateAndParseJWTAuthToken(tokenString, []byte(cfg.SecretKey))
			}

			if err != nil {
				newTokenRequired = errors.Is(err, http.ErrNoCookie) ||
					errors.Is(err, auth.ErrUnexpectedSigningMethod) ||
					errors.Is(err, auth.ErrTokenIsNotValid)

				if !newTokenRequired {
					l.Error("cant validate auth token", zap.Error(err))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			if newTokenRequired {
				userID, err = auth.GenerateUserID()
				if err != nil {
					l.Error("cant create user ID", zap.Error(err))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				tokenString, err = auth.GenerateJWTAuthToken(userID, []byte(cfg.SecretKey))
				if err != nil {
					l.Error("cant create auth token", zap.Error(err))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				http.SetCookie(rw, &http.Cookie{
					Name:     "auth_token",
					Value:    tokenString,
					Path:     "/",
					HttpOnly: true,
				})
			}

			if userID == "" {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := auth.ContextWithUserID(r.Context(), userID)
			h.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
