package service

import (
	"crypto/sha256"
	"encoding/base32"

	"github.com/SergeyRG/shortener/internal/repository"
)

func CreateShortURLID(repo repository.RepositoryURL, url string) string {
	var (
		urlID    string
		hash     [32]byte
		addition string = ""
	)

	for {
		hash = sha256.Sum256([]byte(url + addition))
		urlID = string([]byte(base32.StdEncoding.EncodeToString(hash[:]))[:8])

		if v, ok := repo.GetByID(urlID); !ok {
			break
		} else if v == url {
			break
		} else {
			addition += "1"
		}
	}
	return urlID
}
