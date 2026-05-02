package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap"
)

func JSONShortenHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		userID, ok := auth.UserIDFromContext(req.Context())
		if !ok {
			logging.Logger.Error("cant get user id")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()

		logging.Logger.Debug("start decoding json request")

		jr := &request{}
		if err := decoder.Decode(jr); err != nil {
			logging.Logger.Error("cant decode json request", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		logging.Logger.Debug("json request is decoded")

		ID, err := svc.AddShortURL(context.Background(), jr.URL, userID)
		if err != nil && !errors.Is(err, service.ErrConflict) {
			logging.Logger.Error("cant add short URL", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		status := http.StatusCreated
		if errors.Is(err, service.ErrConflict) {
			status = http.StatusConflict
		}

		shortURL, err := svc.MakeShortURLByID(context.Background(), ID)
		if err != nil {
			logging.Logger.Error("cant make short URL", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		r := response{
			Result: shortURL,
		}

		encoder := json.NewEncoder(rw)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(status)
		logging.Logger.Debug("encoding response and sending")
		err = encoder.Encode(r)
		if err != nil {
			logging.Logger.Debug("error encoding response", zap.Error(err))
			return
		}
		logging.Logger.Debug("response sent", zap.Error(err))

	})
}
