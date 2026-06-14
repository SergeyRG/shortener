package events

import (
	"encoding/json"
	"errors"
	"os"
)

type FileRequestAuditor struct {
	file *os.File
}

func NewFileRequestAuditor(filePath string) (*FileRequestAuditor, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileRequestAuditor{
		file: file,
	}, nil
}

func (fra *FileRequestAuditor) handleAuditEvent(e Event) error {
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

func (fra *FileRequestAuditor) onEvent(e Event) error {
	return fra.handleAuditEvent(e)
}
