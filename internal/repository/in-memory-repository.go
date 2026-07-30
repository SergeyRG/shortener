package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/SergeyRG/shortener/internal/model"
)

// generate:reset
type InMemoryRepositoryURL struct {
	stor     map[string]*model.ShortenModel
	filePath string
	file     *os.File
	mutex    sync.Mutex
}

func NewInMemoryRepositoryURL(data map[string]*model.ShortenModel, filePath string) (*InMemoryRepositoryURL, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return &InMemoryRepositoryURL{
			stor:     data,
			filePath: filePath,
			file:     file,
		}, nil
	}
	return &InMemoryRepositoryURL{
		stor:     make(map[string]*model.ShortenModel),
		filePath: filePath,
		file:     file,
	}, nil
}

func (r *InMemoryRepositoryURL) Add(ctx context.Context, url string, id string, userID string) error {
	if _, ok := r.stor[id]; ok {
		return ErrAlreadyExist
	}

	entry := map[string]*model.ShortenModel{id: {
		ID:          id,
		OriginURL:   url,
		UserID:      userID,
		DeletedFlag: false}}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = r.file.Write(append(([]byte(data)), '\n'))
	if err != nil {
		return err
	}
	r.mutex.Lock()
	r.stor[id] = entry[id]
	r.mutex.Unlock()
	return nil
}

func (r *InMemoryRepositoryURL) GetByID(ctx context.Context, id string) (model.ShortenModel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if v, ok := r.stor[id]; ok {
		return *v, nil
	}
	return model.ShortenModel{}, ErrURLNotFound

}

func (r *InMemoryRepositoryURL) GetByUserID(ctx context.Context, userID string) (url []model.ShortenModel, err error) {
	var result []model.ShortenModel
	r.mutex.Lock()
	defer r.mutex.Unlock()
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
	r.mutex.Lock()
	defer r.mutex.Unlock()
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

func (r *InMemoryRepositoryURL) DeleteBatch(data []model.DeleteTaskDto) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, task := range data {
		for _, id := range task.IDs {
			if r.stor[id].UserID == task.UserID {
				r.stor[id].DeletedFlag = true
			}
		}
	}
	return nil
}

func (r *InMemoryRepositoryURL) Close() error {
	return r.file.Close()
}

func (r *InMemoryRepositoryURL) GetURLSCount(ctx context.Context) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return len(r.stor), nil
	}
}

func (r *InMemoryRepositoryURL) GetUsersCount(ctx context.Context) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	count := 0
	alreadyСounted := make(map[string]bool, len(r.stor))
	for _, v := range r.stor {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			if !alreadyСounted[v.UserID] {
				count++
				alreadyСounted[v.UserID] = true
			}
		}
	}
	return count, nil
}
