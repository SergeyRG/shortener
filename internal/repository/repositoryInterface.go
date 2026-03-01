package repository

type RepositoryURL interface {
	Add(url string, id string)
	GetByID(id string) (url string, result bool)
}
