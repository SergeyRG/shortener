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

	var stor map[string]string
	fileStoreData, err := os.ReadFile(cfg.FileStoragePath)
	if err == nil {
		if unmarshalErr := json.Unmarshal(fileStoreData, &stor); unmarshalErr != nil {
			logger.Error("Ошибка анмаршалинга json. репозиторий будет пустым", zap.Error(err))
			stor = nil
		}
	} else {
		logger.Error("Ошибка чтения файла сохраненных URL. репозиторий будет пустым", zap.Error(err))
	}

	if stor == nil {
		stor = make(map[string]string)
	} else {
		logger.Info("Сохраненные URL успешно загружены")
	}

	repo := repository.NewInMemoryRepositoryURL(stor)
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)

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
