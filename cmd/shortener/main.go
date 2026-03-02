package main

import (
	"net/http"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	repo := repository.NewInMemoryRepositoryURL()
	if err := run(repo); err != nil {
		panic(err)
	}
}

func run(repo repository.RepositoryURL) error {
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handler.RootHandler(repo))
	// mux.HandleFunc("/{id}", handler.RedirectHandler(repo))
	// fmt.Println("Сервер запущен на :8080")
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.RootHandler(repo))
		r.Post("/", handler.RootHandler(repo))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.RedirectHandler(repo))
			r.Post("/", handler.RedirectHandler(repo))
		})
	})
	return http.ListenAndServe(`:8080`, r)
}
