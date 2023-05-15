package txindex

import (
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// TxIndexer //

type (
	startDelegate   func() error
	stopDelegate    func() error
	getTypeDelegate func() string
	indexDelegate   func(types.TxResult) error
)

type mockIndexer struct {
	startFn   startDelegate
	stopFn    stopDelegate
	getTypeFn getTypeDelegate
	indexFn   indexDelegate
}

func (m mockIndexer) Start() error {
	if m.startFn != nil {
		return m.startFn()
	}

	return nil
}

func (m mockIndexer) Stop() error {
	if m.stopFn != nil {
		return m.stopFn()
	}

	return nil
}

func (m mockIndexer) GetType() string {
	if m.getTypeFn != nil {
		return m.getTypeFn()
	}

	return ""
}

func (m mockIndexer) Index(result types.TxResult) error {
	if m.indexFn != nil {
		return m.indexFn(result)
	}

	return nil
}

// EventSwitch //

type (
	fireEventDelegate      func(events.Event)
	addListenerDelegate    func(string, events.EventCallback)
	removeListenerDelegate func(string)
)

type mockEventSwitch struct {
	service.BaseService

	fireEventFn      fireEventDelegate
	addListenerFn    addListenerDelegate
	removeListenerFn removeListenerDelegate
}

func (m *mockEventSwitch) FireEvent(ev events.Event) {
	if m.fireEventFn != nil {
		m.fireEventFn(ev)
	}
}

func (m *mockEventSwitch) AddListener(listenerID string, cb events.EventCallback) {
	if m.addListenerFn != nil {
		m.addListenerFn(listenerID, cb)
	}
}

func (m *mockEventSwitch) RemoveListener(listenerID string) {
	if m.removeListenerFn != nil {
		m.removeListenerFn(listenerID)
	}
}
