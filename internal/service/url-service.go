package service

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/url"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
)

type URLService struct {
	repo        URLRepository
	cfg         config.Config
	idGenerator ShortURLIDGenerator
}

func NewURLService(r URLRepository, cfg config.Config, idGenerator ShortURLIDGenerator) *URLService {
	return &URLService{
		repo:        r,
		cfg:         cfg,
		idGenerator: idGenerator,
	}
}

func (u *URLService) GetOriginalURLByID(ctx context.Context, id string) (string, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *URLService) MakeShortURLByID(ctx context.Context, id string) (string, error) {
	res, err := url.JoinPath(u.cfg.BaseShortURLAddress, id)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (u *URLService) AddShortURL(ctx context.Context, url string) (string, error) {
	var addition = ""
	var id = ""
	for iter := 1; iter <= 10; iter++ {
		id = u.idGenerator.CalculateShortURLID(url + addition)
		err := u.repo.Add(ctx, url, id)

		switch err {
		case nil:
			return id, nil
		//Если такой id уже есть в памяти и url совпадает, то возвращаем этот id
		//Если url не совпадает, то добавляем соль и пересчитываем id
		case repository.ErrAlreadyExist:
			if v, _ := u.repo.GetByID(ctx, id); v == url {
				return id, fmt.Errorf("%w", ErrConflict)
			}
			addition += "1"
		default:
			return "", fmt.Errorf("%w:%w", repository.ErrUnexpected, err)
		}
	}
	return "", repository.ErrNotEnoughID
}

func (u *URLService) AddBatch(ctx context.Context, data []BatchDataRequest) ([]BatchDataResponse, error) {
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
	err := u.repo.AddBatch(ctx, shortURLs)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type URLGenerator struct{}

func (ug URLGenerator) CalculateShortURLID(url string) string {
	var hash = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])
}
