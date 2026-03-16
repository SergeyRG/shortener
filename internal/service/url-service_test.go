package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/config"
	urlErrors "github.com/SergeyRG/shortener/internal/errors"
	"github.com/SergeyRG/shortener/internal/service"
	"github.com/SergeyRG/shortener/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestURLService_GetOriginalURLByID(t *testing.T) {
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
	}
	tests := []struct {
		name    string
		cfg     config.Config
		id      string
		want    string
		wantErr error
	}{
		{
			name:    "check that the value received from the repository is being returned",
			cfg:     cfg,
			id:      "DFSDFDD",
			want:    "http://test.ru",
			wantErr: nil,
		},
		{
			name:    "check that an error is returned if an error has occurred in the repository",
			cfg:     cfg,
			id:      "DFSDFDD",
			want:    "http://test.ru",
			wantErr: errors.New("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLRepository(ctrl)
			m.EXPECT().GetByID(tt.id).Times(1).Return(tt.want, tt.wantErr)
			g := service.URLGenerator{}
			us := service.NewURLService(m, tt.cfg, g)

			got, gotErr := us.GetOriginalURLByID(tt.id)

			if tt.wantErr != nil {
				assert.NotNil(t, gotErr)
			} else {
				assert.Equal(t, got, tt.want)
			}

		})
	}
}

func TestURLService_MakeShortURLByID(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		id      string
		want    string
		wantErr error
	}{
		{
			name: "Check the correctness of the operation when the base URL ends /",
			cfg: config.Config{
				ServerAddress:       ":8080",
				BaseShortURLAddress: "http://test.ru/",
			},
			id:      "DFSDFDD",
			want:    "http://test.ru/DFSDFDD",
			wantErr: nil,
		},
		{
			name: "Check the correctness of the operation when the base URL ends without /",
			cfg: config.Config{
				ServerAddress:       ":8080",
				BaseShortURLAddress: "http://test.ru",
			},
			id:      "DFSDFDD",
			want:    "http://test.ru/DFSDFDD",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLRepository(ctrl)
			g := service.URLGenerator{}
			us := service.NewURLService(m, tt.cfg, g)

			got, gotErr := us.MakeShortURLByID(tt.id)

			if tt.wantErr != nil {
				assert.NotNil(t, gotErr)
			} else {
				assert.Equal(t, got, tt.want)
			}

		})
	}
}

func TestURLService_AddShortURL(t *testing.T) {
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
	}
	tests := []struct {
		name         string
		cfg          config.Config
		url          string
		attempts     int
		repoErr      error
		wantShortURL string
		wantErr      error
	}{
		{
			name:         "1",
			cfg:          cfg,
			url:          "http://test.ru",
			attempts:     1,
			repoErr:      nil,
			wantShortURL: "DSDFDSDF",
			wantErr:      nil,
		},
		{
			name:         "2",
			cfg:          cfg,
			url:          "http://test.ru",
			attempts:     1,
			repoErr:      urlErrors.ErrUnexpected,
			wantShortURL: "DSDFDSDF",
			wantErr:      urlErrors.ErrUnexpected,
		},
		{
			name:         "3",
			cfg:          cfg,
			url:          "http://test.ru",
			attempts:     10,
			repoErr:      urlErrors.ErrAlredyExist,
			wantShortURL: "DSDFDSDF",
			wantErr:      urlErrors.ErrNotEnoughID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mg := mocks.NewMockShortUrlIDGenerator(ctrl)
			mr := mocks.NewMockURLRepository(ctrl)

			mg.EXPECT().CalculateShortURLID(
				gomock.Cond(func(x any) bool { return strings.Contains(x.(string), tt.url) })).
				Times(tt.attempts).Return(tt.wantShortURL)

			mr.EXPECT().Add(tt.url, tt.wantShortURL).Times(tt.attempts).Return(tt.repoErr)
			mr.EXPECT().GetByID(gomock.Any()).AnyTimes().Return("test", nil)

			u := service.NewURLService(mr, tt.cfg, mg)

			got, gotErr := u.AddShortURL(tt.url)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
			} else {
				assert.Equal(t, tt.wantShortURL, got)
			}

		})
	}
}
