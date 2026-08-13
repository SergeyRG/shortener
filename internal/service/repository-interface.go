package service

import (
	"context"

	"github.com/SergeyRG/shortener/internal/model"
)

//go:generate mockgen -destination=mocks/mock-url-reposiyory.gen.go -package=mocks . URLRepository
type URLRepository interface {
	Add(ctx context.Context, url string, id string, userID string) error
	GetByID(ctx context.Context, id string) (url model.ShortenModel, err error)
	GetByUserID(ctx context.Context, userID string) (url []model.ShortenModel, err error)
	Delete(ctx context.Context, id string) error
	AddBatch(ctx context.Context, data []model.ShortenData, userID string) error
	DeleteBatch(data []model.DeleteTaskDto) error
	GetURLSCount(ctx context.Context) (int, error)
	GetUsersCount(ctx context.Context) (int, error)
	Close() error
}
