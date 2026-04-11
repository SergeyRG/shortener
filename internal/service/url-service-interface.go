package service

import "context"

//go:generate mockgen -destination=mocks/mock-url-service.go -package=mocks . URLServiceInterface
type URLServiceInterface interface {
	GetOriginalURLByID(ctx context.Context, id string) (url string, err error)
	AddShortURL(ctx context.Context, url string) (id string, err error)
	MakeShortURLByID(ctx context.Context, id string) (url string, err error)
}
