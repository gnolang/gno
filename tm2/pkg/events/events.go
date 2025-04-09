// Package events - Pub-Sub in go with event caching
package events

import (
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/service"
)

// All implementors must be amino-encodable.
type Event any

// Eventable is the interface reactors and other modules must export to become
// eventable.
type Eventable interface {
	SetEventSwitch(evsw EventSwitch)
}

type Fireable interface {
	FireEvent(ev Event)
}

type Listenable interface {
	// Multiple callbacks can be registered for a given listenerID.  Events are
	// called back in the order that they were registered with this function.
	AddListener(listenerID string, cb EventCallback)
	// Removes all callbacks that match listenerID.
	RemoveListener(listenerID string)
}

// EventSwitch is the interface for synchronous pubsub, where listeners
// subscribe to certain events and, when an event is fired (see Fireable),
// notified via a callback function.
// All listeners are expected to perform work quickly and not block processing
// of the main event emitter.
type EventSwitch interface {
	service.Service
	Fireable
	Listenable
}

type EventCallback func(event Event)

type listenCell struct {
	listenerID string
	cb         EventCallback
}

// This simple implementation is optimized for few listeners.
// This is faster for few listeners, especially for FireEvent.
type eventSwitch struct {
	service.BaseService

	mtx       sync.RWMutex
	listeners []listenCell
}

func NilEventSwitch() EventSwitch {
	return (*eventSwitch)(nil)
}

func NewEventSwitch() EventSwitch {
	evsw := &eventSwitch{
		listeners: make([]listenCell, 0, 10),
	}
	evsw.BaseService = *service.NewBaseService(nil, "EventSwitch", evsw)
	return evsw
}

func (evsw *eventSwitch) OnStart() error {
	return nil
}

func (evsw *eventSwitch) OnStop() {}

func (evsw *eventSwitch) AddListener(listenerID string, cb EventCallback) {
	evsw.mtx.Lock()
	evsw.listeners = append(evsw.listeners, listenCell{listenerID, cb})
	evsw.mtx.Unlock()
}

func (evsw *eventSwitch) RemoveListener(listenerID string) {
	evsw.mtx.Lock()
	newlisteners := make([]listenCell, 0, len(evsw.listeners))
	for _, cell := range evsw.listeners {
		if cell.listenerID != listenerID {
			newlisteners = append(newlisteners, cell)
		}
	}
	evsw.listeners = newlisteners
	evsw.mtx.Unlock()
}

// FireEvent on a nil switch is a noop, but no other operations are allowed for
// safety.
func (evsw *eventSwitch) FireEvent(event Event) {
	if evsw == nil {
		return
	}
	evsw.mtx.RLock()
	listeners := make([]listenCell, len(evsw.listeners))
	copy(listeners, evsw.listeners)
	evsw.mtx.RUnlock()

	for _, cell := range listeners {
		cell.cb(event)
	}
}

func (evsw *eventSwitch) String() string {
	if evsw == nil {
		return "nil-eventSwitch"
	} else {
		evsw.mtx.RLock()
		defer evsw.mtx.RUnlock()

		return fmt.Sprintf("*eventSwitch{%v}", len(evsw.listeners))
	}
}
