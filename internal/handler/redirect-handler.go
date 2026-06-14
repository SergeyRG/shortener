package handler

import (
	"net/http"

	"github.com/SergeyRG/shortener/internal/service"
)

// RedirectHandler возвращает обработчик для перенаправления на оригинальный URL
// по короткой ссылке.
//
// Идентификатор передаются как URL параметр.
func RedirectHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		id := req.PathValue("id")
		url, err := svc.GetOriginalURLByID(req.Context(), id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if url.DeletedFlag {
			rw.WriteHeader(http.StatusGone)
			return
		}
		rw.Header().Set("Location", url.OriginURL)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	}
}
