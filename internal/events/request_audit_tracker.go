package events

import "errors"

type RequestAuditTracker struct {
	observers []AuditObserver
}

func NewRequestAuditTracker() *RequestAuditTracker {
	return &RequestAuditTracker{
		observers: make([]AuditObserver, 0),
	}
}

func (rat *RequestAuditTracker) Register(o AuditObserver) {
	rat.observers = append(rat.observers, o)
}

func (rat *RequestAuditTracker) Notify(e Event) error {
	var errs []error

	for _, o := range rat.observers {
		err := o.Handle(e)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
