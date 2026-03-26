package service

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"net/url"
	"os"

	"github.com/SergeyRG/shortener/internal/config"
	urlErrors "github.com/SergeyRG/shortener/internal/errors"
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

func (u *URLService) ExportRepoToJSONFile() error {
	bytes, err := json.Marshal(u.repo.GetALL())
	if err != nil {
		return err
	}
	f, err := os.OpenFile(u.cfg.FileStoragePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes)

	if err != nil {
		return err
	}

	return nil
}

type URLGenerator struct{}

func (ug URLGenerator) CalculateShortURLID(url string) string {
	var hash = sha256.Sum256([]byte(url))
	return string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])

}
