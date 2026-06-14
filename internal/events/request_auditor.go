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
	TS     time.Time  `json:"ts"`
	Action ActionType `json:"action"`
	UserID string     `json:"user_id"`
	URL    string     `json:"url"`
}

func (e *EventRequestHandled) MarshalJSON() ([]byte, error) {
	type alias EventRequestHandled
	return json.Marshal(&struct {
		TS int64 `json:"ts"`
		alias
	}{
		TS:    e.TS.Unix(),
		alias: alias(*e),
	})
}

func (e *EventRequestHandled) eventType() eventType {
	return EventTypeRequestHandled
}
