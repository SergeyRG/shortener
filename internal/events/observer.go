package events

type eventType int

const (
	EventTypeRequestHandled eventType = iota
)

type Event interface {
	eventType() eventType
}

type AuditObserver interface {
	Handle(Event) error
}

type Subject interface {
	Register(AuditObserver)
	Notify(Event) error
}
