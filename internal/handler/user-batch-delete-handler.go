package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SergeyRG/shortener/internal/auth"
	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/service"
	"go.uber.org/zap"
)

// UserBatchDeleteHandler возвращает обработчик для удаления группы коротких URL
//
// Перечень коротких URL для удаления передается, как json массив.
// Пользователь может удалить только свои URL.
func UserBatchDeleteHandler(svc service.URLServiceInterface) http.HandlerFunc {
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

		jr := []string{}
		if err := decoder.Decode(&jr); err != nil {
			logging.Logger.Error("cant decode json request", zap.Error(err))
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		svc.AddForDeleting(req.Context(), model.DeleteTaskDto{
			UserID: userID,
			IDs:    jr,
		})
		rw.WriteHeader(http.StatusAccepted)
	})
}
