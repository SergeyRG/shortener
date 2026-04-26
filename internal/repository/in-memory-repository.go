package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SergeyRG/shortener/internal/model"
)

type InMemoryRepositoryURL struct {
	stor     map[string]model.ShortenModel
	filePath string
}

func NewInMemoryRepositoryURL(data map[string]model.ShortenModel, filePath string) *InMemoryRepositoryURL {
	if data != nil {
		return &InMemoryRepositoryURL{
			stor:     data,
			filePath: filePath,
		}
	}
	return &InMemoryRepositoryURL{
		stor:     make(map[string]model.ShortenModel),
		filePath: filePath,
	}
}

func (r *InMemoryRepositoryURL) Add(ctx context.Context, url string, id string, userID string) error {
	if _, ok := r.stor[id]; ok {
		return ErrAlreadyExist
	}
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	entry := map[string]model.ShortenModel{id: {
		ID:          id,
		OriginURL:   url,
		UserID:      userID,
		DeletedFlag: false}}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.Write(append(([]byte(data)), '\n'))
	if err != nil {
		return err
	}
	r.stor[id] = entry[id]
	return nil
}

func (r *InMemoryRepositoryURL) GetByID(ctx context.Context, id string) (model.ShortenModel, error) {
	if v, ok := r.stor[id]; ok {
		return v, nil
	} else {
		return model.ShortenModel{}, ErrURLNotFound
	}
}

func (r *InMemoryRepositoryURL) GetByUserID(ctx context.Context, userID string) (url []model.ShortenModel, err error) {
	var result []model.ShortenModel
	for k, v := range r.stor {
		if v.UserID == userID {
			result = append(result, model.ShortenModel{
				ID:          k,
				OriginURL:   v.OriginURL,
				UserID:      userID,
				DeletedFlag: v.DeletedFlag,
			})
		}
	}

	return result, nil
}

func (r *InMemoryRepositoryURL) Delete(ctx context.Context, id string) error {
	delete(r.stor, id)
	return nil
}

func (r *InMemoryRepositoryURL) AddBatch(ctx context.Context, data []model.ShortenData, userID string) error {
	for _, v := range data {
		err := r.Add(ctx, v.OriginURL, v.ID, userID)
		if err != nil {
			return fmt.Errorf("ошибка добавления сокращенного URL для %s:%w", v.OriginURL, err)
		}
	}
	return nil
}
