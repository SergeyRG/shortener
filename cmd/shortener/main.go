package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ошибка запуска приложения: %v", err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("ошибка валидации конфигурации: %s", err)
	}
	if err := logging.Initialize("debug"); err != nil {
		log.Fatalf("ошибка инициализации системы логирования: %s", err)
	}

	logger := logging.Logger
	logger.Info("система логгирования инициализирована, начало инициализации приложения.")
	logger.Info("загрузка сохраненных сокращенных URL")

	var repo service.URLRepository
	var db *sql.DB
	if cfg.DBDSN == "" {
		fileStore, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			logger.Fatal("не удалось открыть/создать файл", zap.Error(err))
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
			logger.Info("Сохраненные URL успешно загружены",
				zap.Int("count", len(stor)))
		} else {
			logger.Info("Файл сохраненных URL пустой, либо произошла ошибка чтения")
		}
		repo = repository.NewInMemoryRepositoryURL(stor)

	} else {
		var err error
		db, err = sql.Open("pgx", cfg.DBDSN)
		if err != nil {
			logger.Fatal("не удалось подключиться к БД", zap.Error(err),
				zap.String("DSN", cfg.DBDSN))
		}
		defer db.Close()
	}

	ps := repository.NewFileRepositoryURL(cfg.FileStoragePath)
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, ps, cfg, g)

	rootHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RootHandler(svc)))
	redirectHandler := logging.WithLogging(middleware.GzipMiddleware(handler.RedirectHandler(svc)))
	JSONShortenHandler := logging.WithLogging(middleware.GzipMiddleware((handler.JSONShortenHandler(svc))))
	CommandHandler := logging.WithLogging(middleware.GzipMiddleware((handler.CommandHandler(svc))))
	DBPingHandler := logging.WithLogging(middleware.GzipMiddleware((handler.DBPingHandler(db))))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", rootHandler)
		r.Get("/{id}", redirectHandler)
		r.Get("/{id}/", redirectHandler)
		r.Get("/ping", DBPingHandler)
		r.Get("/ping/", DBPingHandler)
		r.Post("/api/shorten", JSONShortenHandler)
		r.Post("/command", CommandHandler)
	})
	logger.Info("запуск приложения")
	return http.ListenAndServe(cfg.ServerAddress, r)
}
