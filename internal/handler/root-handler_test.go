package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/handler"
	"github.com/SergeyRG/shortener/internal/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRootHandler(t *testing.T) {
	tests := []struct {
		name            string
		inURL           string
		outID           string
		contentType     string
		wantLocation    string
		wantErr         error
		wantContentType string
	}{
		{
			name:            "test the return of the short URL",
			inURL:           "http:/test.ru",
			outID:           "JDGRTDSN",
			contentType:     "text/plain",
			wantLocation:    "http://shortURL",
			wantErr:         nil,
			wantContentType: "text/plain",
		},
		{
			name:            "test the return the error on bad content-type",
			inURL:           "http:/test.ru",
			outID:           "JDGRTDSN",
			contentType:     "application/json",
			wantLocation:    "http://shortURL",
			wantErr:         errors.New(""),
			wantContentType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLServiceInterface(ctrl)
			if tt.wantErr == nil {
				m.EXPECT().AddShortURL(gomock.Any(), tt.inURL, "test").Times(1).Return(tt.outID, tt.wantErr)

				m.EXPECT().MakeShortURLByID(gomock.Any(), tt.outID).Times(1).Return(tt.wantLocation, tt.wantErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.inURL))
			req.Header.Set("Content-Type", tt.contentType)
			rw := httptest.NewRecorder()

			rootHandler := handler.RootHandler(m)

			r := chi.NewRouter()
			r.Post("/", rootHandler)

			r.ServeHTTP(rw, req)

			res := rw.Result()
			defer res.Body.Close()

			if tt.wantErr != nil {
				assert.Equal(t, res.StatusCode, http.StatusBadRequest)
			} else {
				assert.Equal(t, tt.wantContentType, res.Header.Get("content-type"))
				assert.Equal(t, http.StatusCreated, res.StatusCode)
				assert.Equal(t, tt.wantLocation, rw.Body.String())
			}
		})
	}
}
