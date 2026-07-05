package service

//go:generate mockgen -destination=mocks/mock-short-URL-ID-generator.go -package=mocks . ShortURLIDGenerator
type ShortURLIDGenerator interface {
	CalculateShortURLID(url string) string
}
