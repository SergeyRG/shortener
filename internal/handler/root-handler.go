package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/SergeyRG/shortener/internal/service"
)

func RootHandler(svc service.URLServiceInterface) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {

		log.Printf("Content-type: %s\n", req.Header.Get("content-type"))
		log.Printf("url: %s\n", req.URL.Path)
		log.Printf("method: %s\n", req.Method)

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

		log.Printf("Body: %s\n", body)
		url := string(body)

		id, err := svc.AddShortURL(url)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusCreated)

		shortUrl, err := svc.MakeShortURLByID(id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Write([]byte(shortUrl))

	}
}
