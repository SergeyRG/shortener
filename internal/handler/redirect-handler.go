package handler

import (
	"log"
	"net/http"

	"github.com/SergeyRG/shortener/internal/service"
)

func RedirectHandler(svc service.URLServiceInterface) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {

		log.Printf("Content-type: %s\n", req.Header.Get("content-type"))
		log.Printf("url: %s\n", req.URL.Path)
		log.Printf("method: %s\n", req.Method)

		id := req.PathValue("id")
		url, err := svc.GetOriginalURLByID(id)
		log.Printf("url: %s id: %s\n", url, id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		rw.Header().Set("Location", url)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	}
}
