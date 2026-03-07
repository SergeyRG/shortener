package service

//go:generate mockgen -destination=mocks/mock-ShortURLIDGenerator.go -package=mocks . ShortURLIDGenerator
type ShortURLIDGenerator interface {
	CalculateShortURLID(url string) string
}
