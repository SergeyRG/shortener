package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap/zaptest"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func Test_rootHandler(t *testing.T) {
	logging.Logger = zaptest.NewLogger(t)

	tmpFile, _ := os.CreateTemp("", "test_*.tmp")
	tmpFile.Close()
	repo, err := repository.NewInMemoryRepositoryURL(nil, tmpFile.Name())
	if err != nil {
		t.Fatalf("cant create repo: %v", err)
	}
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
		SecretKey:           "test",
	}
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)
	authMiddleware := middleware.Auth(cfg)
	h := authMiddleware(handler.RootHandler(svc))
	srv := httptest.NewServer(h)

	defer srv.Close()

	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	tests := []struct {
		name        string
		target      string
		method      string
		body        string
		contentType string
		want        want
	}{
		{
			name:        `Post request to "/" returns a short link`,
			target:      "/",
			method:      http.MethodPost,
			body:        `http://ya.ru`,
			contentType: `text/plain`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: `text/plain`,
				body:        `http://localhost:8080/HGHQZJH6`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := resty.New().SetBaseURL(srv.URL).R()
			resp, err := req.SetHeader("Content-Type", tt.contentType).
				SetBody(tt.body).
				Execute(tt.method, tt.target)

			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Contains(t, resp.Header().Get("Content-Type"), tt.want.contentType)

			assert.Equal(t, tt.want.body, strings.TrimSpace(string(resp.Body())))
		})
	}
}

func Test_redirectHandler(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.tmp")
	if err != nil {
		t.Fatal("Cant create temp file")
	}
	tmpFile.Close()
	repo, err := repository.NewInMemoryRepositoryURL(nil, tmpFile.Name())
	if err != nil {
		t.Fatalf("cant create repo: %v", err)
	}
	repo.Add(context.Background(), "http://ya.ru", "HGHQZJH6", "test")
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
	}
	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)
	h := handler.RedirectHandler(svc)
	r := chi.NewRouter()
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	type want struct {
		statusCode int
		body       string
		location   string
	}
	tests := []struct {
		name   string
		target string
		method string
		body   string
		want   want
	}{
		{
			name:   `Get request to "/{id}" returns a redirection`,
			target: "/HGHQZJH6",
			method: http.MethodGet,
			body:   ``,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				body:       ``,
				location:   `http://ya.ru`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := resty.New().SetBaseURL(srv.URL).
				SetRedirectPolicy(resty.NoRedirectPolicy()).
				R()

			resp, err := req.SetBody(tt.body).
				Execute(tt.method, tt.target)

			if err != nil && !errors.Is(err, resty.ErrAutoRedirectDisabled) {
				t.Fatalf("error making HTTP request")
			}

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())

			assert.Equal(t, tt.want.body, strings.TrimSpace(string(resp.Body())))

			if tt.want.location != "" {
				assert.Equal(t, resp.Header().Get("location"), tt.want.location)
			}
		})
	}
}
