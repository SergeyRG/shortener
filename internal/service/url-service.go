package service

import (
	"crypto/sha256"
	"encoding/base32"
	"net/url"

	"github.com/SergeyRG/shortener/internal/config"
	urlErrors "github.com/SergeyRG/shortener/internal/errors"
)

type URLService struct {
	repo        URLRepository
	cfg         config.Config
	idGenerator ShortUrlIDGenerator
}

func NewURLService(r URLRepository, cfg config.Config, idGenerator ShortUrlIDGenerator) *URLService {
	return &URLService{
		repo:        r,
		cfg:         cfg,
		idGenerator: idGenerator,
	}
}

func (u *URLService) GetOriginalURLByID(id string) (string, error) {
	if v, err := u.repo.GetByID(id); err != nil {
		return "", err
	} else {
		return v, nil
	}
}

func (u *URLService) MakeShortURLByID(id string) (string, error) {
	res, err := url.JoinPath(u.cfg.BaseShortURLAddress, id)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (u *URLService) AddShortURL(url string) (string, error) {
	var addition = ""
	var id = ""
	var iter = 1
	for {
		id = u.idGenerator.CalculateShortURLID(url + addition)
		err := u.repo.Add(url, id)

		if err == nil {
			return id, nil
		}

		if err != urlErrors.ErrAlredyExist {
			return "", urlErrors.ErrUnexpected
		}

		if v, _ := u.repo.GetByID(id); v == url {
			return id, nil
		}

		addition += "1"
		iter += 1

		if iter == 11 {
			return "", urlErrors.ErrNotEnoughID
		}
	}
}

type UrlGenerator struct{}

func (ug UrlGenerator) CalculateShortURLID(url string) string {
	var hash = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])

}
