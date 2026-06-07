package service

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/url"
	"time"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
)

type URLService struct {
	repo        URLRepository
	cfg         config.Config
	idGenerator ShortURLIDGenerator
	deleteCh    chan model.DeleteTaskDto
}

func NewURLService(r URLRepository, cfg config.Config, idGenerator ShortURLIDGenerator) *URLService {
	us := &URLService{
		repo:        r,
		cfg:         cfg,
		idGenerator: idGenerator,
		deleteCh:    make(chan model.DeleteTaskDto, 1),
	}
	go us.startDeleteWorker()
	return us
}

func (u *URLService) GetOriginalURLByID(ctx context.Context, id string) (model.ShortenModel, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *URLService) GetURLByUserID(ctx context.Context, userID string) ([]UserURLData, error) {
	userURL, err := u.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []UserURLData

	for _, url := range userURL {
		data := UserURLData{}
		data.ShortURL, err = u.MakeShortURLByID(ctx, url.ID)
		if err != nil {
			return nil, err
		}
		data.OriginalURL = url.OriginURL

		result = append(result, data)
	}

	return result, nil
}

func (u *URLService) MakeShortURLByID(ctx context.Context, id string) (string, error) {
	res, err := url.JoinPath(u.cfg.BaseShortURLAddress, id)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (u *URLService) AddShortURL(ctx context.Context, url string, userID string) (string, error) {
	var addition = ""
	var id = ""
	for iter := 1; iter <= 10; iter++ {
		id = u.idGenerator.CalculateShortURLID(url + addition)
		err := u.repo.Add(ctx, url, id, userID)

		switch err {
		case nil:
			return id, nil
		//Если такой id уже есть в памяти и url совпадает, то возвращаем этот id
		//Если url не совпадает, то добавляем соль и пересчитываем id
		case repository.ErrAlreadyExist:
			if v, _ := u.repo.GetByID(ctx, id); v.OriginURL == url {
				return id, fmt.Errorf("%w", ErrConflict)
			}
			addition += "1"
		default:
			return "", fmt.Errorf("%w:%w", repository.ErrUnexpected, err)
		}
	}
	return "", repository.ErrNotEnoughID
}

func (u *URLService) AddBatch(ctx context.Context, data []BatchDataRequest, userID string) ([]BatchDataResponse, error) {
	shortURLs := make([]model.ShortenData, len(data))
	response := make([]BatchDataResponse, len(data))
	for i, v := range data {
		shortURLID := u.idGenerator.CalculateShortURLID(v.OriginalURL)
		shortURL, err := u.MakeShortURLByID(ctx, shortURLID)
		if err != nil {
			return nil, err
		}
		shortURLs[i] = model.ShortenData{
			ID:        shortURLID,
			OriginURL: v.OriginalURL,
		}
		response[i] = BatchDataResponse{
			CorrelationID: v.CorrelationID,
			ShortURL:      shortURL,
		}
	}
	err := u.repo.AddBatch(ctx, shortURLs, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (u *URLService) AddForDeleting(ctx context.Context, data model.DeleteTaskDto) {
	u.deleteCh <- data
}

func (u *URLService) startDeleteWorker() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	var buf []model.DeleteTaskDto

	for {
		select {
		case task, ok := <-u.deleteCh:
			if !ok {
				if len(buf) > 0 {
					u.repo.DeleteBatch(buf)
				}
				return
			}
			buf = append(buf, task)
			if len(buf) == 1000 {
				u.repo.DeleteBatch(buf)
				buf = buf[:0]
			}
		case <-ticker.C:
			if len(buf) > 0 {
				u.repo.DeleteBatch(buf)
				buf = buf[:0]
			}
		}
	}
}

type URLGenerator struct{}

func (ug URLGenerator) CalculateShortURLID(url string) string {
	var hash = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])
}
