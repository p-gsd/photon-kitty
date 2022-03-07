package events

import lua "github.com/yuin/gopher-lua"

var registry = make(EventCallbacks)

type EventCallback func(e Event) error

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
	EventTypeArticleOpened   = EventType("ArticleOpened")
	EventTypeLinkOpened      = EventType("LinkOpened")
)

type Init struct{}

func (e *Init) Type() EventType {
	return EventTypeInit
}

type RunMediaStart struct {
	Link string
	Card func(*lua.LState) lua.LValue
}

func (e *RunMediaStart) Type() EventType {
	return EventTypeRunMediaStart
}

type RunMediaEnd struct {
	Link string
	Card func(*lua.LState) lua.LValue
}

func (e *RunMediaEnd) Type() EventType {
	return EventTypeRunMediaEnd
}

type FeedsDownloaded struct{}

func (e *FeedsDownloaded) Type() EventType {
	return EventTypeFeedsDownloaded
}

type ArticleOpened struct {
	Link string
	Card func(*lua.LState) lua.LValue
}

func (e *ArticleOpened) Type() EventType {
	return EventTypeArticleOpened
}

type LinkOpened struct {
	Link string
	Card func(*lua.LState) lua.LValue
}

func (e *LinkOpened) Type() EventType {
	return EventTypeLinkOpened
}
