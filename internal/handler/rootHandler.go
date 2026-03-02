package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
)

func RootHandler(repo repository.RepositoryURL, cfg config.Config) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {

		fmt.Printf("Content-type: %s\n", req.Header.Get("content-type"))
		fmt.Printf("url: %s\n", req.URL.Path)
		fmt.Printf("method: %s\n", req.Method)

		contentType := req.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}
		if req.Method != http.MethodPost {
			http.Error(rw, "Only POST is allowed", http.StatusBadRequest)
			return
		}
		if req.URL.Path != "/" {
			http.Error(rw, "URL is not allowed", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		fmt.Printf("Body: %s\n", body)
		url := string(body)
		urlID := service.CreateShortURLID(repo, url)
		repo.Add(url, urlID)

		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusCreated)

		rw.Write([]byte(cfg.BaseShortUrlAddress + `/` + urlID))

	}
}
