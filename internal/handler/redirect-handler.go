package handler

import (
	"net/http"

	"github.com/SergeyRG/shortener/internal/service"
)

func RedirectHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		id := req.PathValue("id")
		url, err := svc.GetOriginalURLByID(req.Context(), id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		rw.Header().Set("Location", url)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	}
}
