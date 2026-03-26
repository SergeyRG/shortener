package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
)

func CommandHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		contentType := req.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		switch string(body) {
		case "SaveURLs":
			logging.Logger.Info("Получена команда сохранения URL")
			svc.ExportRepoToJSONFile()
			rw.WriteHeader(http.StatusCreated)
			return
		default:
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
