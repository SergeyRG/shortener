package events

import (
	"encoding/json"
	"time"
)

type ActionType string

const (
	ActionTypeUnknown ActionType = "unknown"
	ActionTypeShorten ActionType = "shorten"
	ActionTypeFollow  ActionType = "follow"
)

type RequestAuditor interface {
	handleAuditEvent(Event) error
}

type EventRequestHandled struct {
	Ts     time.Time  `json:"ts"`
	Action ActionType `json:"action"`
	UserID string     `json:"user_id"`
	URL    string     `json:"url"`
}

func (e *EventRequestHandled) MarshalJSON() ([]byte, error) {
	type alias EventRequestHandled
	return json.Marshal(&struct {
		Ts int64 `json:"ts"`
		alias
	}{
		Ts:    e.Ts.Unix(),
		alias: alias(*e),
	})
}

func (e *EventRequestHandled) eventType() eventType {
	return EventTypeRequestHandled
}
