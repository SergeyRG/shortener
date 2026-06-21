package events

import (
	"context"
	"time"

	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
)

type RequestAuditor struct {
	EventCh chan Event
	timeOut time.Duration
	rat     *RequestAuditTracker
}

func NewRequestAuditor(rat *RequestAuditTracker, EventCh chan Event, t time.Duration) *RequestAuditor {
	if rat == nil {
		return nil
	}
	return &RequestAuditor{
		rat:     rat,
		timeOut: t,
		EventCh: EventCh,
	}
}

func (ra *RequestAuditor) SendEvent(e Event) {
	ctx, cancel := context.WithTimeout(context.Background(), ra.timeOut)
	defer cancel()

	select {
	case ra.EventCh <- e:
	case <-ctx.Done():
		logging.Logger.Error(
			"Событие аудита не записано в канал аудита. Запись в канал прервана по таймауту",
			zap.Any("event", e))
	}
}

func (ra *RequestAuditor) StartAuditTracking() {
	logging.Logger.Debug("Запуск потока аудита запросов")
	go func() {
		for e := range ra.EventCh {
			ra.rat.Notify(e)
		}
	}()
}
