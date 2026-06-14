package events

import "errors"

type RequestAuditTracker struct {
	observers []Observer
}

func NewRequestAuditTracker() *RequestAuditTracker {
	return &RequestAuditTracker{
		observers: make([]Observer, 0),
	}
}

func (rat *RequestAuditTracker) Register(o Observer) {
	rat.observers = append(rat.observers, o)
}

func (rat *RequestAuditTracker) Notify(e Event) error {
	var errs []error

	for _, o := range rat.observers {
		err := o.onEvent(e)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
