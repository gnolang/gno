package server

import (
	"github.com/gnolang/gnostats/proto"
	"github.com/rs/xid"
)

type (
	// dataStream is the active hub -> dashboard data stream
	dataStream chan *proto.DataPoint

	// subs is the subscription store,
	// which maps the subscription ID -> stream
	subs map[xid.ID]dataStream
)

// subscribe creates a new data stream subscription
func (s subs) subscribe() (xid.ID, dataStream) {
	var (
		id = xid.New()
		ch = make(dataStream, 1)
	)

	s[id] = ch

	return id, ch
}

// unsubscribe removes the given subscription
func (s subs) unsubscribe(id xid.ID) {
	if ch := s[id]; ch != nil {
		// Close the notification channel
		close(ch)
	}

	// Delete the subscription
	delete(s, id)
}

// notify notifies all subscription listeners
func (s subs) notify(data *proto.DataPoint) {
	// Notify the listeners
	for _, ch := range s {
		select {
		case ch <- data:
		default:
		}
	}
}
