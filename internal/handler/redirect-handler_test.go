package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"go.uber.org/mock/gomock"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name         string
		inID         string
		outURL       model.ShortenModel
		wantLocation string
		wantErr      error
	}{
		{"test the return of the redirect", "NMGHFDBG",
			model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
			"http:/test.ru", nil},
		{"test the return of the error on non existen id", "NMGHFDBG",
			model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
			"", repository.ErrAlreadyExist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLServiceInterface(ctrl)
			m.EXPECT().GetOriginalURLByID(gomock.Any(), tt.inID).Times(1).Return(tt.outURL, tt.wantErr)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.inID, nil)
			rw := httptest.NewRecorder()

			redirectHandler := handler.RedirectHandler(m)

			r := chi.NewRouter()
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", redirectHandler)
			})
			r.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			if tt.wantErr != nil {
				assert.Equal(t, res.StatusCode, http.StatusBadRequest)
			} else {
				assert.Equal(t, res.Header.Get("Location"), tt.wantLocation)
				assert.Equal(t, res.StatusCode, http.StatusTemporaryRedirect)
			}
		})
	}
}
