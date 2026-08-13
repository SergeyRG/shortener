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

			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				l.Error("cant get value from cookies", zap.Error(err))
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err == nil {
				tokenString = cookieToken.Value
			}

			userID, err := auth.ProccessToken(tokenString, []byte(cfg.SecretKey))

			var e auth.ErrNewTokenRequerd
			if errors.As(err, &e) {
				http.SetCookie(rw, &http.Cookie{
					Name:     "auth_token",
					Value:    e.NewToken,
					Path:     "/",
					HttpOnly: true,
				})
				err = nil
			}

			if err != nil {
				l.Error("auth error", zap.Error(err))
				rw.WriteHeader(http.StatusInternalServerError)
				return
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
