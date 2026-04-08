package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

	if err := logging.Initialize("debug"); err != nil {
		log.Fatalf("Ошибка инициализации системы логирования: %s", err)
	}
	logger := logging.Logger
	logger.Info("система логгирования инициализирована, начало инициализации приложения.")

	logger.Info("загрузка сохраненных сокращенных URL")

	fileStore, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatal("Не удалось открыть/создать файл", zap.Error(err))
	}
	defer fileStore.Close()

	decoder := json.NewDecoder(fileStore)
	stor := make(map[string]string)

	for decoder.More() {
		if err := decoder.Decode(&stor); err != nil {
			logging.Logger.Error("Ошибка восстановления сохраненных URL", zap.Error(err))
			stor = make(map[string]string)
			break
		}
	}
	if len(stor) > 0 {
		logger.Info("Сохраненные URL успешно загружены", zap.Int("count", len(stor)))
	} else {
		logger.Info("Файл сохраненных URL пустой, либо произошла ошибка чтения")
	}

	repo := repository.NewInMemoryRepositoryURL(stor)
	ps := repository.NewFileRepositoryURL(cfg.FileStoragePath)
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, ps, cfg, g)

	rootHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RootHandler(svc)))
	redirectHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RedirectHandler(svc)))
	JSONShortenHandler := logging.WithLogging(middleware.GzipMiddleware((handler.JSONShortenHandler(svc))))
	CommandHandler := logging.WithLogging(middleware.GzipMiddleware((handler.CommandHandler(svc))))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", rootHandler)
		r.Get("/{id}", redirectHandler)
		r.Get("/{id}/", redirectHandler)
		r.Post("/api/shorten", JSONShortenHandler)
		r.Post("/command", CommandHandler)
	})
	logger.Info("запуск приложения")
	return http.ListenAndServe(cfg.ServerAddress, r)
}
