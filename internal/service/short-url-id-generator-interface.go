package service

//go:generate mockgen -destination=mocks/mock-ShortUrlIDGenerator.go -package=mocks . ShortUrlIDGenerator
type ShortUrlIDGenerator interface {
	CalculateShortURLID(url string) string
}
