package repository

import (
	urlErrors "github.com/SergeyRG/shortener/internal/errors"
)

type InMemoryRepositoryURL struct {
	stor map[string]string
}

func NewInMemoryRepositoryURL(data map[string]string) *InMemoryRepositoryURL {
	if data != nil {
		return &InMemoryRepositoryURL{
			stor: data,
		}
	}
	return &InMemoryRepositoryURL{
		stor: make(map[string]string),
	}
}

func (r *InMemoryRepositoryURL) Add(url string, id string) error {
	if _, ok := r.stor[id]; ok {
		return urlErrors.ErrAlredyExist
	}
	r.stor[id] = url
	return nil
}

func (r *InMemoryRepositoryURL) GetByID(id string) (string, error) {
	if v, ok := r.stor[id]; ok {
		return v, nil
	} else {
		return "", urlErrors.ErrURLNotFound
	}

}

func (r *InMemoryRepositoryURL) GetALL() map[string]string {
	return r.stor
}
