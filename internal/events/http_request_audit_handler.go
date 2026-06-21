package events

import (
	"fmt"
	"time"

	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/go-resty/resty/v2"
)

type HTTPRequestAuditHandler struct {
	URL string
}

func NewHTTPRequestAuditHandler(URL string) *HTTPRequestAuditHandler {
	return &HTTPRequestAuditHandler{
		URL: URL,
	}
}

func (hra *HTTPRequestAuditHandler) handleAuditEvent(e Event) error {
	logging.Logger.Debug("Передача события аудита запросов на web server")
	HTTPClient := resty.New().
		SetTimeout(5*time.Second).
		SetHeader("Content-type", "application/json")

	resp, err := HTTPClient.NewRequest().Post(hra.URL)
	if err != nil {
		return fmt.Errorf("ошибка выполнения http запроса: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf(
			"ошибка обработки события на сервере, http код: %v", resp.StatusCode())
	}
	return nil
}

func (hra *HTTPRequestAuditHandler) onEvent(e Event) error {
	return hra.handleAuditEvent(e)
}
