package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_rootHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	tests := []struct {
		name        string
		target      string
		method      string
		body        io.Reader
		contentType string
		want        want
	}{
		{
			name:        `Post request to "/" returns a short link`,
			target:      "/",
			method:      http.MethodPost,
			body:        strings.NewReader(`http://ya.ru`),
			contentType: `text/plain`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: `text/plain`,
				body:        `http://localhost:8080/HGHQZJH6`,
			},
		},
		{
			name:        `GET request to "/" returns a bad request`,
			target:      "/",
			method:      http.MethodGet,
			body:        strings.NewReader(``),
			contentType: `text/plain`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: `text/plain`,
				body:        `Only POST is allowed`,
			},
		},
		{
			name:        `POST request for urls other than "/" and "/{id}" returns a bad request`,
			target:      "/test/test",
			method:      http.MethodPost,
			body:        strings.NewReader(`ya.ru`),
			contentType: `text/plain`,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: `text/plain`,
				body:        `URL is not allowed`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewInMemoryRepositoryURL()
			mux := http.NewServeMux()
			mux.HandleFunc("/", handler.RootHandler(repo))

			req := httptest.NewRequest(tt.method, tt.target, tt.body)
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			result := rec.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), tt.want.contentType)

			urlResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.body, strings.TrimSpace(string(urlResult)))
		})
	}
}

func Test_redirectHandler(t *testing.T) {
	type existedURL struct {
		url string
		id  string
	}
	type want struct {
		statusCode int
		body       string
		location   string
	}
	tests := []struct {
		name       string
		target     string
		method     string
		body       io.Reader
		existedURL existedURL
		want       want
	}{
		{
			name:   `Get request to "/{id}" returns a redirection`,
			target: "/HGHQZJH6",
			method: http.MethodGet,
			body:   strings.NewReader(``),
			existedURL: existedURL{
				url: `http://ya.ru`,
				id:  `HGHQZJH6`,
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				body:       ``,
				location:   `http://ya.ru`,
			},
		},
		{
			name:   `Post request to "/{id}" returns a bad request`,
			target: "/HGHQZJH6",
			method: http.MethodPost,
			body:   strings.NewReader(``),
			existedURL: existedURL{
				url: `http://ya.ru`,
				id:  `HGHQZJH6`,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				body:       `Bad request`,
				location:   ``,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewInMemoryRepositoryURL()
			repo.Add(tt.existedURL.url, tt.existedURL.id)
			mux := http.NewServeMux()
			mux.HandleFunc("/{id}", handler.RedirectHandler(repo))

			req := httptest.NewRequest(tt.method, tt.target, tt.body)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			result := rec.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			urlResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.body, strings.TrimSpace(string(urlResult)))

			if tt.want.location != "" {
				assert.Equal(t, result.Header.Get("location"), tt.want.location)
			}
		})
	}
}
