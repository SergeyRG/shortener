package service

import (
	"context"

	"github.com/SergeyRG/shortener/internal/model"
)

type BatchDataRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchDataResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//go:generate mockgen -destination=mocks/mock-url-service.go -package=mocks . URLServiceInterface
type URLServiceInterface interface {
	GetOriginalURLByID(ctx context.Context, id string) (url model.ShortenModel, err error)
	AddShortURL(ctx context.Context, url string, userID string) (id string, err error)
	MakeShortURLByID(ctx context.Context, id string) (url string, err error)
	AddBatch(ctx context.Context, data []BatchDataRequest, userID string) ([]BatchDataResponse, error)
	GetURLByUserID(ctx context.Context, userID string) (url []UserURLData, err error)
	AddForDeleting(ctx context.Context, data model.DeleteTaskDto)
}
