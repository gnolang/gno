package dial

import (
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

// Pop removes an item from the dial queue, if any
func (q *Queue) Pop() *Item {
	q.mux.Lock()
	defer q.mux.Unlock()

	return q.items.PopFront()
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
