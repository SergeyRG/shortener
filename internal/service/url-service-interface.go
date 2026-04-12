package service

import "context"

type BatchDataRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchDataResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//go:generate mockgen -destination=mocks/mock-url-service.go -package=mocks . URLServiceInterface
type URLServiceInterface interface {
	GetOriginalURLByID(ctx context.Context, id string) (url string, err error)
	AddShortURL(ctx context.Context, url string) (id string, err error)
	MakeShortURLByID(ctx context.Context, id string) (url string, err error)
	AddBatch(ctx context.Context, data []BatchDataRequest) ([]BatchDataResponse, error)
}
