package emitter

import (
	"sync"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
)

// ServerEmitter is an `emitter.Emitter`
var _ emitter.Emitter = (*ServerEmitter)(nil)

type ServerEmitter struct {
	events   []events.Event
	muEvents sync.Mutex
}

func (m *ServerEmitter) Emit(evt events.Event) {
	m.muEvents.Lock()
	defer m.muEvents.Unlock()

	m.events = append(m.events, evt)
}

func (m *ServerEmitter) NextEvent() (evt events.Event) {
	m.muEvents.Lock()
	defer m.muEvents.Unlock()

	if len(m.events) > 0 {
		// pull next event from the list
		evt, m.events = m.events[0], m.events[1:]
	}

	return evt
}
