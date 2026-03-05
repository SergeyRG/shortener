package service

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/url"

	"github.com/SergeyRG/shortener/internal/config"
	urlErrors "github.com/SergeyRG/shortener/internal/errors"
)

type URLService struct {
	repo RepositoryURL
	cfg  config.Config
}

func NewURLService(r RepositoryURL, cfg config.Config) *URLService {
	return &URLService{
		repo: r,
		cfg:  cfg,
	}
}

func (u *URLService) GetOriginalURLByID(id string) (string, error) {
	if v, err := u.repo.GetByID(id); err != nil {
		return "", err
	} else {
		return v, nil
	}
}

func (u *URLService) MakeShortURLByID(id string) string {
	res, _ := url.JoinPath(u.cfg.BaseShortURLAddress, id)
	return res
}

func (u *URLService) AddShortURL(url string) (string, error) {
	var addition = ""
	var id = ""
	for {
		id = calculateShortURLID(url + addition)
		err := u.repo.Add(url, id)

		if err == nil {
			return id, nil
		}

		if err != urlErrors.ErrAlredyExist {
			return "", errors.New("unexpected error")
		}

		if v, _ := u.repo.GetByID(id); v == url {
			return id, nil
		}

		addition += "1"
	}
}

func calculateShortURLID(url string) string {
	var hash [32]byte = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])

}
