package events

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/SergeyRG/shortener/internal/logging"
)

type FileRequestAuditHandler struct {
	file *os.File
}

func NewFileRequestAuditHandler(filePath string) (*FileRequestAuditHandler, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileRequestAuditHandler{
		file: file,
	}, nil
}

func (fra *FileRequestAuditHandler) handleAuditEvent(e Event) error {
	logging.Logger.Debug("Запись события аудита запросов в файл")
	if e.eventType() != EventTypeRequestHandled {
		return nil
	}
	e, ok := e.(*EventRequestHandled)
	if !ok {
		return errors.New("ошибка привидения типа события")
	}
	eventJSON, err := json.Marshal(e)
	if err != nil {
		return errors.New("ошибка преобразования события в JSON формат")
	}
	eventJSON = append(eventJSON, '\n')

	_, err = fra.file.Write(eventJSON)
	if err != nil {
		return errors.New("ошибка записи события в файл")
	}
	return nil
}

func (fra *FileRequestAuditHandler) onEvent(e Event) error {
	return fra.handleAuditEvent(e)
}
