package handler

import (
	"net/http"

	"github.com/SergeyRG/shortener/internal/events"
)

type AuditableHandler func(rw http.ResponseWriter, r *http.Request) events.Event

func WithAudit(ra *events.RequestAuditor, h AuditableHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		e := h(rw, r)
		if ra == nil {
			return
		}
		if e != nil {
			ra.SendEvent(e)
		}
	}
}
