package events

type eventType int

const (
	EventTypeRequestHandled eventType = iota
)

type Event interface {
	eventType() eventType
}

type Observer interface {
	onEvent(Event) error
}

type Subject interface {
	Register(Observer)
	Notify(Event) error
}
