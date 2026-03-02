package handler

import (
	"fmt"
	"net/http"

	"github.com/SergeyRG/shortener/internal/repository"
)

func RedirectHandler(repo repository.RepositoryURL) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {

		fmt.Printf("Content-type: %s\n", req.Header.Get("content-type"))
		fmt.Printf("url: %s\n", req.URL.Path)
		fmt.Printf("method: %s\n", req.Method)

		if req.Method != http.MethodGet {

			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		id := req.PathValue("id")
		url, ok := repo.GetByID(id)
		fmt.Printf("url: %s id: %s\n", url, id)
		if !ok {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		rw.Header().Set("Location", url)
		rw.WriteHeader(http.StatusTemporaryRedirect)
	}
}
