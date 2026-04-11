package service

import "context"

//go:generate mockgen -destination=mocks/mock-url-reposiyory.go -package=mocks . URLRepository
type URLRepository interface {
	Add(ctx context.Context, url string, id string) error
	GetByID(ctx context.Context, id string) (url string, err error)
	Delete(ctx context.Context, id string) error
}
