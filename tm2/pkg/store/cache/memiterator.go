package cache

import (
	"container/list"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      []*std.KVPair
	ascending  bool
	cur        int
}

func newMemIterator(start, end []byte, items *list.List, ascending bool) *memIterator {
	itemsInDomain := make([]*std.KVPair, 0)
	var entered bool
	for e := items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*std.KVPair)
		if !dbm.IsKeyInDomain(item.Key, start, end) {
			if entered {
				break
			}
			continue
		}
		itemsInDomain = append(itemsInDomain, item)
		entered = true
	}

	cur := 0
	if !ascending {
		cur = len(itemsInDomain) - 1
	}

	return &memIterator{
		start:     start,
		end:       end,
		items:     itemsInDomain,
		ascending: ascending,
		cur:       cur,
	}
}

func (mi *memIterator) Domain() ([]byte, []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Valid() bool {
	if mi.cur < 0 || len(mi.items) <= mi.cur {
		return false
	}
	return true
}

func (mi *memIterator) assertValid() {
	if !mi.Valid() {
		panic("memIterator is invalid")
	}
}

func (mi *memIterator) Next() {
	mi.assertValid()
	if mi.ascending {
		mi.cur++
	} else {
		mi.cur--
	}
}

func (mi *memIterator) Key() []byte {
	mi.assertValid()
	return mi.items[mi.cur].Key
}

func (mi *memIterator) Value() []byte {
	mi.assertValid()
	return mi.items[mi.cur].Value
}

func (mi *memIterator) Close() {
	mi.start = nil
	mi.end = nil
	mi.items = nil
}
