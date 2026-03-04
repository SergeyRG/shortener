package service

type RepositoryURL interface {
	Add(url string, id string) error
	GetByID(id string) (url string, err error)
}
