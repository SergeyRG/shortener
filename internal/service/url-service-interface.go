package service

//go:generate mockgen -destination=mocks/mock-url-service.go -package=mocks . URLServiceInterface
type URLServiceInterface interface {
	GetOriginalURLByID(id string) (url string, err error)
	AddShortURL(url string) (id string, err error)
	MakeShortURLByID(id string) (url string, err error)
	ExportRepoToJSONFile() error
}
