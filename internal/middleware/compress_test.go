package middleware_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/middleware"
	"github.com/SergeyRG/shortener/internal/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGzipCompression(t *testing.T) {

	t.Run("sends_gzip", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		m := mocks.NewMockURLServiceInterface(ctrl)
		m.EXPECT().AddShortURL(gomock.Any()).AnyTimes()
		m.EXPECT().MakeShortURLByID(gomock.Any()).AnyTimes()

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte("http://test.ru"))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		handler := middleware.GzipMiddleware(handler.RootHandler(m))

		req := httptest.NewRequest(http.MethodPost, "/", buf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Accept-Encoding", "")

		rw := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Post("/", handler)

		r.ServeHTTP(rw, req)

		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rw.Result().StatusCode)

	})
}
