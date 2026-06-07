package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/SergeyRG/shortener/migrations"
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

		decoder := json.NewDecoder(fileStore)
		stor := make(map[string]*model.ShortenModel)
		for decoder.More() {
			if err := decoder.Decode(&stor); err != nil {
				logging.Logger.Error("Ошибка восстановления сохраненных URL", zap.Error(err))
				stor = make(map[string]*model.ShortenModel)
				break
			}
		}
		if len(stor) > 0 {
			logger.Info("Сохраненные URL успешно загружены",
				zap.Int("count", len(stor)))
		} else {
			logger.Info("Файл сохраненных URL пустой, либо произошла ошибка чтения")
		}
		fileStore.Close()

		repo, err = repository.NewInMemoryRepositoryURL(stor, cfg.FileStoragePath)
		if err != nil {
			logger.Fatal("не удалось создать объект репозиториия", zap.Error(err))
		}
	} else {
		var err error
		db, err = sql.Open("pgx", cfg.DBDSN)
		if err != nil {
			logger.Fatal("не удалось подключиться к БД", zap.Error(err),
				zap.String("DSN", cfg.DBDSN))
		}

		defer db.Close()
		repo, err = repository.NewPSQLDBRepositoryURL(db)
		if err != nil {
			logger.Fatal("не удалось инициализировать репозиторий", zap.Error(err))
		}
		err = migrations.RunMigrations(db)
		if err != nil {
			logger.Fatal("не удалось мигрировать БД", zap.Error(err))
		}
	}

	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)

	r := initRouter(svc, db, cfg)
	logger.Info("запуск приложения")

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	defer repo.Close()
	return server.ListenAndServe()
}

func initRouter(svc service.URLServiceInterface, db *sql.DB, cfg config.Config) chi.Router {
	rootHandler := handler.RootHandler(svc)
	redirectHandler := handler.RedirectHandler(svc)
	JSONShortenHandler := handler.JSONShortenHandler(svc)
	DBPingHandler := handler.DBPingHandler(db)
	BatchAddHandler := handler.BatchAddHandler(svc)
	UserURLHandler := handler.UserURLHandler(svc)
	UserBatchDeleteHandler := handler.UserBatchDeleteHandler(svc)

	authMiddleware := middleware.Auth(cfg)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(logging.WithLogging)
		r.Use(authMiddleware)
		r.Use(middleware.GzipMiddleware)
		r.Post("/", rootHandler)
		r.Get("/{id}", redirectHandler)
		r.Get("/{id}/", redirectHandler)
		r.Get("/ping", DBPingHandler)
		r.Get("/ping/", DBPingHandler)
		r.Post("/api/shorten", JSONShortenHandler)
		r.Post("/api/shorten/batch", BatchAddHandler)
		r.Get("/api/user/urls", UserURLHandler)
		r.Delete("/api/user/urls", UserBatchDeleteHandler)
	})
	return r
}
