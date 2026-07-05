package handler

import (
	"net/http"
	"time"

	"github.com/SergeyRG/shortener/internal/events"
	"github.com/SergeyRG/shortener/internal/service"
)

// RedirectHandler возвращает обработчик для перенаправления на оригинальный URL
// по короткой ссылке.
//
// Идентификатор передаются как URL параметр.
func RedirectHandler(svc service.URLServiceInterface) AuditableHandler {
	return func(rw http.ResponseWriter, req *http.Request) events.Event {

		id := req.PathValue("id")
		url, err := svc.GetOriginalURLByID(req.Context(), id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return nil
		}
		if url.DeletedFlag {
			rw.WriteHeader(http.StatusGone)
			return nil
		}
		rw.Header().Set("Location", url.OriginURL)
		rw.WriteHeader(http.StatusTemporaryRedirect)

		return &events.EventRequestHandled{
			TS:     time.Now(),
			Action: events.ActionTypeFollow,
			URL:    url.OriginURL,
		}
	}
}
