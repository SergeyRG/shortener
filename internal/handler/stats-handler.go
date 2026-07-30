package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap"
)

type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

func StatsHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		urlsCount, err := svc.GetURLSCount(req.Context())
		if err != nil {
			logging.Logger.Error("cant get short URLs count", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		usersCount, err := svc.GetUsersCount(req.Context())
		if err != nil {
			logging.Logger.Error("cant get users count", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		encoder := json.NewEncoder(rw)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		logging.Logger.Debug("encoding response and sending")

		err = encoder.Encode(statsResponse{
			URLs:  urlsCount,
			Users: usersCount,
		})
		if err != nil {
			logging.Logger.Debug("error encoding response", zap.Error(err))
			return
		}
		logging.Logger.Debug("response sent")
	})
}
