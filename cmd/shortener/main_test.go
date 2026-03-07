package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func Test_rootHandler(t *testing.T) {
	repo := repository.NewInMemoryRepositoryURL()
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
	}
	g := service.UrlGenerator{}
	svc := service.NewURLService(repo, cfg, g)
	h := http.HandlerFunc(handler.RootHandler(svc))
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
	repo := repository.NewInMemoryRepositoryURL()
	repo.Add("http://ya.ru", "HGHQZJH6")
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
	}
	g := service.UrlGenerator{}
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
