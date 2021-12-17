package events

var registry = make(EventCallbacks)

type EventCallback func(Event) error

type EventCallbacks map[EventType][]EventCallback

func (ec EventCallbacks) Subscribe(et EventType, c EventCallback) {
	ec[et] = append(ec[et], c)
}

func (ec EventCallbacks) Emit(e Event) {
	for _, c := range ec[e.Type()] {
		c(e)
	}
}

func Emit(e Event) {
	registry.Emit(e)
}

type Event interface {
	Type() EventType
}

func Subscribe(et EventType, c EventCallback) {
	registry.Subscribe(et, c)
}

type EventType string

const (
	EventTypeInit            = EventType("Init")
	EventTypeRunMediaStart   = EventType("RunMediaStart")
	EventTypeRunMediaEnd     = EventType("RunMediaEnd")
	EventTypeFeedsDownloaded = EventType("FeedsDownloaded")
)

var events = make(EventCallbacks)

type Init struct{}

func (e *Init) Type() EventType {
	return EventTypeInit
}

type RunMediaStart struct {
	Link string
}

func (e *RunMediaStart) Type() EventType {
	return EventTypeRunMediaStart
}

type RunMediaEnd struct {
	Link string
}

func (e *RunMediaEnd) Type() EventType {
	return EventTypeRunMediaEnd
}

type FeedsDownloaded struct{}

func (e *FeedsDownloaded) Type() EventType {
	return EventTypeFeedsDownloaded
}
