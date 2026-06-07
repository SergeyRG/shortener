package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
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
		want    model.ShortenModel
		wantErr error
	}{
		{
			name:    "check that the value received from the repository is being returned",
			cfg:     cfg,
			id:      "DFSDFDD",
			want:    model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
			wantErr: nil,
		},
		{
			name:    "check that an error is returned if an error has occurred in the repository",
			cfg:     cfg,
			id:      "DFSDFDD",
			want:    model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
			wantErr: errors.New("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLRepository(ctrl)

			m.EXPECT().GetByID(context.Background(), tt.id).Times(1).Return(tt.want, tt.wantErr)

			g := service.URLGenerator{}
			us := service.NewURLService(m, tt.cfg, g)

			got, gotErr := us.GetOriginalURLByID(context.Background(), tt.id)

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

			got, gotErr := us.MakeShortURLByID(context.Background(), tt.id)

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
			repoErr:      fmt.Errorf("Тестовая ошибка"),
			wantShortURL: "DSDFDSDF",
			wantErr:      fmt.Errorf("%w:%w", repository.ErrUnexpected, fmt.Errorf("Тестовая ошибка")),
		},
		{
			name:         "3",
			cfg:          cfg,
			url:          "http://test.ru",
			attempts:     10,
			repoErr:      repository.ErrAlreadyExist,
			wantShortURL: "DSDFDSDF",
			wantErr:      repository.ErrNotEnoughID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mg := mocks.NewMockShortURLIDGenerator(ctrl)
			mr := mocks.NewMockURLRepository(ctrl)

			mg.EXPECT().CalculateShortURLID(
				gomock.Cond(func(x any) bool { return strings.Contains(x.(string), tt.url) })).
				Times(tt.attempts).Return(tt.wantShortURL)

			mr.EXPECT().Add(context.Background(), tt.url, tt.wantShortURL, "test").Times(tt.attempts).Return(tt.repoErr)
			mr.EXPECT().GetByID(gomock.Any(), gomock.Any()).AnyTimes().Return(
				model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false}, nil)

			u := service.NewURLService(mr, tt.cfg, mg)

			got, gotErr := u.AddShortURL(context.Background(), tt.url, "test")
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
			} else {
				assert.Equal(t, tt.wantShortURL, got)
			}

		})
	}
}
