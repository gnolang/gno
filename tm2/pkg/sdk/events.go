package sdk

import (
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

// ----------------------------------------------------------------------------
// EventLogger
// ----------------------------------------------------------------------------

// EventLogger implements a simple wrapper around a slice of Event objects that
// can be emitted from.
type EventLogger struct {
	events []Event
}

func NewEventLogger() *EventLogger {
	return &EventLogger{nil}
}

func (em *EventLogger) Events() []Event { return em.events }

// EmitEvent stores a single Event object.
func (em *EventLogger) EmitEvent(event Event) {
	em.events = append(em.events, event)
}

// EmitEvents stores a series of Event objects.
func (em *EventLogger) EmitEvents(events []Event) {
	em.events = append(em.events, events...)
}

// ----------------------------------------------------------------------------
// Event
// ----------------------------------------------------------------------------

type Event = abci.Event

type EventAttribute struct {
	Key   string
	Value string
}

type AttributedEvent struct {
	Type       string
	Attributes []EventAttribute
}

func NewEvent(ty string, attrs ...EventAttribute) Event {
	return AttributedEvent{Type: ty, Attributes: attrs}
}

func NewEventAttribute(key, value string) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

func (e AttributedEvent) AssertABCIEvent() {}

func (e AttributedEvent) String() string {
	return fmt.Sprintf("type: %s, attributes: %v", e.Type, e.Attributes)
}

func (ea EventAttribute) String() string {
	return fmt.Sprintf("%s: %s", ea.Key, ea.Value)
}
