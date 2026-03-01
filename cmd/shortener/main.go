package main

import (
	"fmt"
	"net/http"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
)

func main() {
	repo := repository.NewInMemoryRepositoryURL()
	if err := run(repo); err != nil {
		panic(err)
	}
}

func run(repo repository.RepositoryURL) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.RootHandler(repo))
	mux.HandleFunc("/{id}", handler.RedirectHandler(repo))
	fmt.Println("Сервер запущен на :8080")
	return http.ListenAndServe(`:8080`, mux)
}
