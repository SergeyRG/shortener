package service

//go:generate mockgen -destination=mocks/mock-url-reposiyory.go -package=mocks . URLRepository
type URLRepository interface {
	Add(url string, id string) error
	GetByID(id string) (url string, err error)
}
