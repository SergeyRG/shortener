package service

import (
	"context"

	"github.com/SergeyRG/shortener/internal/model"
)

//go:generate mockgen -destination=mocks/mock-url-reposiyory.go -package=mocks . URLRepository
type URLRepository interface {
	Add(ctx context.Context, url string, id string, userID string) error
	GetByID(ctx context.Context, id string) (url []string, err error)
	GetByUserID(ctx context.Context, userID string) (url []model.ShortenData, err error)
	Delete(ctx context.Context, id string) error
	AddBatch(ctx context.Context, data []model.ShortenData, userID string) error
}
