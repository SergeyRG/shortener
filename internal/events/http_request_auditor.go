package events

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type HTTPRequestAuditor struct {
	URL string
}

func NewHTTPRequestAuditor(URL string) *HTTPRequestAuditor {
	return &HTTPRequestAuditor{
		URL: URL,
	}
}

func (hra *HTTPRequestAuditor) handleAuditEvent(e Event) error {
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

func (hra *HTTPRequestAuditor) onEvent(e Event) error {
	return hra.handleAuditEvent(e)
}
