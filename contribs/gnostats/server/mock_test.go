package server

import (
	"github.com/gnolang/gnostats/proto"
	"github.com/rs/xid"
)

type (
	subscribeDelegate   func() (xid.ID, dataStream)
	unsubscribeDelegate func(xid.ID)
	notifyDelegate      func(*proto.DataPoint)
)

type mockSubscriptions struct {
	subscribeFn   subscribeDelegate
	unsubscribeFn unsubscribeDelegate
	notifyFn      notifyDelegate
}

func (m *mockSubscriptions) subscribe() (xid.ID, dataStream) {
	if m.subscribeFn != nil {
		return m.subscribeFn()
	}

	return xid.NilID(), nil
}

func (m *mockSubscriptions) unsubscribe(id xid.ID) {
	if m.unsubscribeFn != nil {
		m.unsubscribeFn(id)
	}
}

func (m *mockSubscriptions) notify(data *proto.DataPoint) {
	if m.notifyFn != nil {
		m.notifyFn(data)
	}
}
