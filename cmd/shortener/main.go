package main

import (
	"log"
	"net/http"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v", err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка валидации конфигурации: %s", err)
	}

	if err := logging.Initialize("info"); err != nil {
		log.Fatalf("Ошибка инициализации системы логирования: %s", err)
	}
	logger := logging.Logger
	logger.Info("система логгирования инициализирована, начало инициализации приложения.")

	repo := repository.NewInMemoryRepositoryURL()
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)

	rootHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RootHandler(svc)))
	redirectHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RedirectHandler(svc)))
	JSONShortenHandler := logging.WithLogging(middleware.GzipMiddleware((handler.JSONShortenHandler(svc))))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", rootHandler)
		r.Get("/{id}", redirectHandler)
		r.Get("/{id}/", redirectHandler)
		r.Post("/api/shorten", JSONShortenHandler)
	})
	logger.Info("запуск приложения")
	return http.ListenAndServe(cfg.ServerAddress, r)
}
