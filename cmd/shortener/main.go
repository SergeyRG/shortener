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
	"path"
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

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ошибка запуска приложения: %v", err)
	}
}

func run() error {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	stopCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %s", err)
	}
	if err := logging.Initialize("debug"); err != nil {
		return fmt.Errorf("ошибка инициализации системы логирования: %s", err)
	}

	logger := logging.Logger
	logger.Info("система логгирования инициализирована, начало инициализации приложения.")
	logger.Info("загрузка сохраненных сокращенных URL")

	var repo service.URLRepository
	var db *sql.DB
	if cfg.DBDSN == "" {
		fileStore, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("не удалось открыть/создать файл: %v", err)
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
			return fmt.Errorf("не удалось создать объект репозиториия: %v", err)
		}
	} else {
		var err error
		db, err = sql.Open("pgx", cfg.DBDSN)
		if err != nil {
			return fmt.Errorf("не удалось подключиться к БД: %v, DBDSN: %s", err, cfg.DBDSN)
		}

		defer db.Close()
		repo, err = repository.NewPSQLDBRepositoryURL(db)
		if err != nil {
			return fmt.Errorf("не удалось инициализировать репозиторий: %v", err)
		}
		err = migrations.RunMigrations(db)
		if err != nil {
			return fmt.Errorf("не удалось мигрировать БД: %v", err)
		}
	}

	requestAuditor, err := initRequestAuditor(cfg)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать аудит запросов: %v", err)
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
		var ServerErr error
		if cfg.EnableHTTPS {
			exePath, err := os.Executable()
			exePath = path.Dir(exePath)
			if err != nil {
				return fmt.Errorf("не удалось определить местоположение исполняемого файла: %v", err)
			}
			logger.Info("Запуск сервера в режиме HTTPS")
			ServerErr = server.ListenAndServeTLS(
				path.Join(exePath, "tls", "cert.pem"),
				path.Join(exePath, "tls", "key.pem"),
			)
		} else {
			logger.Info("Запуск сервера в режиме HTTP")
			ServerErr = server.ListenAndServe()
		}
		if ServerErr != nil && !errors.Is(ServerErr, http.ErrServerClosed) {
			return fmt.Errorf("критическая ошибка HTTP-сервера: %w", ServerErr)
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
