package dial

import (
	"slices"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	queue "github.com/sig-0/insertion-queue"
)

// Item is a single dial queue item, wrapping
// the approximately appropriate dial time, and the
// peer dial address
type Item struct {
	Time    time.Time         // appropriate dial time
	Address *types.NetAddress // the dial address of the peer
}

// Less is the comparison method for the dial queue Item (time ascending)
func (i Item) Less(item Item) bool {
	return i.Time.Before(item.Time)
}

// Queue is a time-sorted (ascending) dial queue
type Queue struct {
	mux sync.RWMutex

	items queue.Queue[Item] // sorted dial queue (by time, ascending)
}

// NewQueue creates a new dial queue
func NewQueue() *Queue {
	return &Queue{
		items: queue.NewQueue[Item](),
	}
}

// Peek returns the first item in the dial queue, if any
func (q *Queue) Peek() *Item {
	q.mux.RLock()
	defer q.mux.RUnlock()

	if q.items.Len() == 0 {
		return nil
	}

	item := q.items.Index(0)

	return &item
}

// Push adds new items to the dial queue
func (q *Queue) Push(items ...Item) {
	q.mux.Lock()
	defer q.mux.Unlock()

	for _, item := range items {
		q.items.Push(item)
	}
}

// Pop removes an item from the dial queue, if any. The dial loop pops through
// PopDue instead, so a removal between its peek and pop cannot hand it an item
// that is not due
func (q *Queue) Pop() *Item {
	q.mux.Lock()
	defer q.mux.Unlock()

	return q.items.PopFront()
}

// PopDue removes and returns the first item in the dial queue, if it is due at
// the given time. The check and the removal happen under one lock, so a
// concurrent Remove cannot slip an item that is not due to the head in between
func (q *Queue) PopDue(now time.Time) *Item {
	q.mux.Lock()
	defer q.mux.Unlock()

	if q.items.Len() == 0 || now.Before(q.items.Index(0).Time) {
		return nil
	}

	return q.items.PopFront()
}

// Remove removes every item dialing the given address from the dial queue
func (q *Queue) Remove(addr *types.NetAddress) {
	q.mux.Lock()
	defer q.mux.Unlock()

	q.items = slices.DeleteFunc(q.items, func(i Item) bool {
		return addr.Equals(*i.Address)
	})
}

// Has returns a flag indicating if the given
// address is in the dial queue
func (q *Queue) Has(addr *types.NetAddress) bool {
	q.mux.RLock()
	defer q.mux.RUnlock()

	for _, i := range q.items {
		if addr.Equals(*i.Address) {
			return true
		}
	}

	return false
}
