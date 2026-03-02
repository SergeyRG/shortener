package main

import (
	"net/http"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()
	repo := repository.NewInMemoryRepositoryURL()
	if err := run(repo, cfg); err != nil {
		panic(err)
	}
}

func run(repo repository.RepositoryURL, cfg config.Config) error {
	rootHandler := handler.RootHandler(repo, cfg)
	redirectHandler := handler.RedirectHandler(repo)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", rootHandler)
		r.Post("/", rootHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", redirectHandler)
			r.Post("/", redirectHandler)
		})
	})
	return http.ListenAndServe(cfg.ServerAddress, r)
}
