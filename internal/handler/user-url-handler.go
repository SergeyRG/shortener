package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SergeyRG/shortener/internal/auth"
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

// UserURLHandler возвращает обработчик для получения коротких URL созданных
// пользователем.
//
// Пользователь может получить, только короткие URL, ссозданный им.
// В случае успеха, возвращается массив структур в виде json:
//
//	 [
//			{
//			"short_url":"DFFGFDSD"
//			"original_url":"http://test.test"
//			}
//	 ]
func UserURLHandler(svc service.URLServiceInterface) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		userID, ok := auth.UserIDFromContext(req.Context())
		if !ok {
			logging.Logger.Error("cant get user id")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		userURL, err := svc.GetURLByUserID(req.Context(), userID)
		if err != nil {
			logging.Logger.Error("cant add short URL", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		status := http.StatusOK
		if len(userURL) == 0 {
			status = http.StatusNoContent
		}

		encoder := json.NewEncoder(rw)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(status)

		logging.Logger.Debug("encoding response and sending")
		err = encoder.Encode(userURL)
		if err != nil {
			logging.Logger.Debug("error encoding response", zap.Error(err))
			return
		}
		logging.Logger.Debug("response sent", zap.Error(err))

	})
}
