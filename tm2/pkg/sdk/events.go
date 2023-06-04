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

func NewEvent(ty string, attrs ...string) *AttributedEvent {
	if len(attrs)%2 == 1 {
		attrs = append(attrs, "")
	}

	eventAttrs := make([]EventAttribute, 0, len(attrs)/2)
	for i := 0; i < len(attrs); i += 2 {
		attr := EventAttribute{Key: attrs[i], Value: attrs[i+1]}
		eventAttrs = append(eventAttrs, attr)
	}

	return &AttributedEvent{Type: ty, Attributes: eventAttrs}
}

func NewEventAttribute(key, value string) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

func (e AttributedEvent) AssertABCIEvent() {}

func (e *AttributedEvent) AddAttribute(key, value string) {
	e.Attributes = append(e.Attributes, EventAttribute{Key: key, Value: value})
}

func (e AttributedEvent) String() string {
	return fmt.Sprintf("type: %s, attributes: %v", e.Type, e.Attributes)
}

func (ea EventAttribute) String() string {
	return fmt.Sprintf("%s: %s", ea.Key, ea.Value)
}
