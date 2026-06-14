package handler

import (
	"database/sql"
	"net/http"

	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
)

// DBPingHandler возвращает обработчик для проверки соединения с БД.
//
// Обрабатывает GET запросы.
// Возвращает ответ с кодом 200 в случае успешного установления соедининия
// с БД.
// Возвращает ответ с кодом 506 в случае ошибки.
func DBPingHandler(db *sql.DB) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if db == nil {
			rw.WriteHeader(http.StatusInsufficientStorage)
			return
		}
		if err := db.PingContext(req.Context()); err != nil {
			logging.Logger.Debug("ошибка подключения к БД",
				zap.Error(err))
			rw.WriteHeader(http.StatusInsufficientStorage)
			return
		} else {
			logging.Logger.Debug("подключение к БД успешно установлено",
				zap.Error(err))
			rw.WriteHeader(http.StatusOK)
			return
		}
	}
}
