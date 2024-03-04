package emitter

import (
	"sync"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
)

type ServerMock struct {
	events   []events.Event
	muEvents sync.Mutex
}

func (m *ServerMock) Emit(evt events.Event) {
	m.muEvents.Lock()
	m.events = append(m.events, evt)
	m.muEvents.Unlock()
}

func (m *ServerMock) NextEvent() (evt events.Event) {
	m.muEvents.Lock()
	if len(m.events) > 0 {
		// pull next event from the list
		evt, m.events = m.events[0], m.events[1:]
	}
	m.muEvents.Unlock()

	return evt
}
