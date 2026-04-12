package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SergeyRG/shortener/internal/model"
)

type InMemoryRepositoryURL struct {
	stor     map[string]string
	filePath string
}

func NewInMemoryRepositoryURL(data map[string]string, filePath string) *InMemoryRepositoryURL {
	if data != nil {
		return &InMemoryRepositoryURL{
			stor:     data,
			filePath: filePath,
		}
	}
	return &InMemoryRepositoryURL{
		stor:     make(map[string]string),
		filePath: filePath,
	}
}

func (r *InMemoryRepositoryURL) Add(ctx context.Context, url string, id string) error {
	if _, ok := r.stor[id]; ok {
		return ErrAlredyExist
	}
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	entry := map[string]string{id: url}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.Write(append(([]byte(data)), '\n'))
	if err != nil {
		return err
	}
	r.stor[id] = url
	return nil
}

func (r *InMemoryRepositoryURL) GetByID(ctx context.Context, id string) (string, error) {
	if v, ok := r.stor[id]; ok {
		return v, nil
	} else {
		return "", ErrURLNotFound
	}

}

func (r *InMemoryRepositoryURL) Delete(ctx context.Context, id string) error {
	delete(r.stor, id)
	return nil
}

func (r *InMemoryRepositoryURL) AddBatch(ctx context.Context, data []model.ShortenData) error {
	for _, v := range data {
		err := r.Add(ctx, v.OriginURL, v.ID)
		if err != nil {
			return fmt.Errorf("ошибка добавления сокращенного URL для %s:%w", v.OriginURL, err)
		}
	}
	return nil
}
