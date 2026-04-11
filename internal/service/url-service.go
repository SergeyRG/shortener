package service

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/url"

	"github.com/SergeyRG/shortener/internal/config"
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
		case repository.ErrAlredyExist:
			if v, _ := u.repo.GetByID(ctx, id); v == url {
				return id, nil
			}
			addition += "1"
		default:
			return "", fmt.Errorf("%w:%w", repository.ErrUnexpected, err)
		}
	}
	return "", repository.ErrNotEnoughID
}

type URLGenerator struct{}

func (ug URLGenerator) CalculateShortURLID(url string) string {
	var hash = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])

}
