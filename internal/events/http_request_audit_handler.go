package events

import (
	"fmt"
	"time"

	"github.com/SergeyRG/shortener/internal/logging"
	"github.com/go-resty/resty/v2"
)

type HTTPRequestAuditHandler struct {
	URL        string
	HTTPClient *resty.Client
}

func NewHTTPRequestAuditHandler(URL string) *HTTPRequestAuditHandler {
	HTTPClient := resty.New().
		SetTimeout(5*time.Second).
		SetHeader("Content-type", "application/json")

	return &HTTPRequestAuditHandler{
		URL:        URL,
		HTTPClient: HTTPClient,
	}
}

func (hra *HTTPRequestAuditHandler) handleAuditEvent(e Event) error {
	logging.Logger.Debug("Передача события аудита запросов на web server")

	resp, err := hra.HTTPClient.NewRequest().SetBody(e).Post(hra.URL)
	if err != nil {
		return fmt.Errorf("ошибка выполнения http запроса: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf(
			"ошибка обработки события на сервере, http код: %v", resp.StatusCode())
	}
	return nil
}

func (hra *HTTPRequestAuditHandler) Handle(e Event) error {
	return hra.handleAuditEvent(e)
}
