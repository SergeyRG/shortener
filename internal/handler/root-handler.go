package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
)

// RootHandler возвращает обработчик для создания короткого URL, при этом оригинальный
// URL передаетс в виде простого текста в body. Например:
//
// В случае успеха, возвращает идентификатор также в виде текста в body.
func RootHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		userID, ok := auth.UserIDFromContext(req.Context())
		if !ok {
			logging.Logger.Error("cant get user id")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		contentType := req.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		url := string(body)
		id, err := svc.AddShortURL(req.Context(), url, userID)

		if err != nil && !errors.Is(err, service.ErrConflict) {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
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
			return
		}
		rw.Write([]byte(shortURL))
	}
}
