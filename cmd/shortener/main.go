package main

import (
	"log"
	"net/http"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка валидации конфигурации: %s", err)
	}
	repo := repository.NewInMemoryRepositoryURL()
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)
	if err := run(svc, cfg); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v", err)
	}
}

func run(svc service.URLServiceInterface, cfg config.Config) error {
	rootHandler := handler.RootHandler(svc)
	redirectHandler := handler.RedirectHandler(svc)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", rootHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", redirectHandler)
		})
	})
	return http.ListenAndServe(cfg.ServerAddress, r)
}
