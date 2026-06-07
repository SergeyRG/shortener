package middleware

import (
	"net/http"
	"time"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/events"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func WithRequestAudit(rt *events.RequestAuditTracker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(rw, r)

			if rt == nil {
				return
			}

			action := events.ActionTypeUnknown

			if (r.URL.Path == "/api/shorten" || r.URL.Path == "/") && r.Method == "POST" {
				action = events.ActionTypeShorten
			}

			routePattern := ""
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				routePattern = rctx.RoutePattern()
			}

			if routePattern == "/{id}" && r.Method == "GET" {
				action = events.ActionTypeFollow
			}

			if action == events.ActionTypeUnknown {
				return
			}

			userID, _ := auth.UserIDFromContext(r.Context())

			e := events.EventRequestHandled{
				TS:     time.Now(),
				Action: action,
				UserID: userID,
				URL:    r.URL.String(),
			}

			err := rt.Notify(&e)

			if err != nil {
				logging.Logger.Error(
					"произошла ошибка обработки события аудита запроса", zap.Error(err))
			}
		})
	}
}
