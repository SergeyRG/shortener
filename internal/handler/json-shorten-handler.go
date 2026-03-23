package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap"
)

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

func JSONShortenHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()

		logging.Logger.Debug("start decoding json request")

		jr := &request{}
		if err := decoder.Decode(jr); err != nil {
			logging.Logger.Debug("cant decode json request", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		logging.Logger.Debug("json request is decoded")

		ID, err := svc.AddShortURL(jr.URL)
		if err != nil {
			logging.Logger.Debug("cant add make short URL", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, err := svc.MakeShortURLByID(ID)
		if err != nil {
			logging.Logger.Debug("cant add make short URL", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		r := response{
			Result: shortURL,
		}

		encoder := json.NewEncoder(rw)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		logging.Logger.Debug("encoding response and sending")
		err = encoder.Encode(r)
		if err != nil {
			logging.Logger.Debug("error encoding response", zap.Error(err))
			return
		}
		logging.Logger.Debug("response is encoded", zap.Error(err))

	})
}
