package service

type URLServiceInterface interface {
	GetOriginalURLByID(id string) (url string, err error)
	AddShortURL(url string) (id string, err error)
	MakeShortURLByID(id string) string
}
