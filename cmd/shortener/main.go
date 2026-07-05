package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/events"
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
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ошибка запуска приложения: %v", err)
	}
}

func run() error {
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	requestAuditor, err := initRequestAuditor(cfg)
	if err != nil {
		logger.Fatal("не удалось инициализировать аудит запросов", zap.Error(err))
	}

	if requestAuditor != nil {
		requestAuditor.StartAuditTracking()
	}

	errg, ctx := errgroup.WithContext(stopCtx)

	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)

	errg.Go(func() error {
		svc.StartDeleteWorker(ctx)
		return nil
	})

	r := initRouter(svc, db, cfg, requestAuditor)
	logger.Info("запуск приложения")

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	errg.Go(func() error {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("критическая ошибка HTTP-сервера: %w", err)
		}
		return nil
	})

	errg.Go(func() error {
		<-ctx.Done()
		logger.Debug("начало остановки приложения")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("ошибка при остановке HTTP-сервера", zap.Error(err))
		}

		svc.Close()
		return nil
	})

	if err := errg.Wait(); err != nil {
		return fmt.Errorf("приложение завершилось с ошибкой: %w", err)
	}

	logger.Info("Приложение успешно остановлено")
	return nil

}

func initRouter(
	svc service.URLServiceInterface,
	db *sql.DB,
	cfg config.Config,
	ra *events.RequestAuditor,
) chi.Router {
	rootHandlerAuditable := handler.RootHandler(svc)
	rootHandler := handler.WithAudit(ra, rootHandlerAuditable)

	redirectHandlerAuditable := handler.RedirectHandler(svc)
	redirectHandler := handler.WithAudit(ra, redirectHandlerAuditable)

	JSONShortenHandlerAuditable := handler.JSONShortenHandler(svc)
	JSONShortenHandler := handler.WithAudit(ra, JSONShortenHandlerAuditable)

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

func initRequestAuditor(cfg config.Config) (*events.RequestAuditor, error) {
	if cfg.AuditFilePath == "" && cfg.AuditURL == "" {
		return nil, nil
	}

	rt := events.NewRequestAuditTracker()

	if cfg.AuditFilePath != "" {
		fa, err := events.NewFileRequestAuditHandler(cfg.AuditFilePath)
		if err != nil {
			return nil, err
		}
		rt.Register(fa)
	}
	if cfg.AuditURL != "" {
		ha := events.NewHTTPRequestAuditHandler(cfg.AuditURL)
		rt.Register(ha)
	}

	ra := events.NewRequestAuditor(rt, make(chan events.Event, 1), time.Second*10)
	return ra, nil
}
