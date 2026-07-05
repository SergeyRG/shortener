package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/events"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
)

// RootHandler возвращает обработчик для создания короткого URL, при этом оригинальный
// URL передаетс в виде простого текста в body. Например:
//
// В случае успеха, возвращает идентификатор также в виде текста в body.
func RootHandler(svc service.URLServiceInterface) AuditableHandler {
	return func(rw http.ResponseWriter, req *http.Request) events.Event {
		userID, ok := auth.UserIDFromContext(req.Context())
		if !ok {
			logging.Logger.Error("cant get user id")
			rw.WriteHeader(http.StatusInternalServerError)
			return nil
		}
		contentType := req.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return nil
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return nil
		}

		url := string(body)
		id, err := svc.AddShortURL(req.Context(), url, userID)

		if err != nil && !errors.Is(err, service.ErrConflict) {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return nil
		}

		status := http.StatusCreated
		if errors.Is(err, service.ErrConflict) {
			status = http.StatusConflict
		}

		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(status)

		shortURL, err := svc.MakeShortURLByID(req.Context(), id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return nil
		}
		rw.Write([]byte(shortURL))

		return &events.EventRequestHandled{
			TS:     time.Now(),
			Action: events.ActionTypeShorten,
			UserID: userID,
			URL:    url,
		}
	}
}
